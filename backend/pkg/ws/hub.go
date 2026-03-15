package ws

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	QuizCode        string
	QuizTitle       string
	TimePerQuestion int
	SelfPaced       bool
	HostUserID      string // quiz creator — only they can advance questions
	clients         map[*Client]bool
	broadcast       chan []byte
	register        chan *Client
	unregister      chan *Client
	questions       []QuestionData
	currentQ        int
	answeredUsers   map[string]bool // tracks who answered current question (live mode)
	scorer          Scorer
	leaderboard     LeaderboardProvider
	mu              sync.RWMutex
}

func NewHub(quizCode, quizTitle string, timePerQuestion int, selfPaced bool, hostUserID string, questions []QuestionData, scorer Scorer, leaderboard LeaderboardProvider) *Hub {
	return &Hub{
		QuizCode:        quizCode,
		QuizTitle:       quizTitle,
		TimePerQuestion: timePerQuestion,
		SelfPaced:       selfPaced,
		HostUserID:      hostUserID,
		clients:         make(map[*Client]bool),
		broadcast:       make(chan []byte, 256),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		questions:       questions,
		currentQ:        -1,
		answeredUsers:   make(map[string]bool),
		scorer:          scorer,
		leaderboard:     leaderboard,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			count := len(h.clients)

			mode := "live"
			if h.SelfPaced {
				mode = "self_paced"
			}

			// Send welcome
			welcome, _ := NewMessage("welcome", WelcomePayload{
				QuizTitle:      h.QuizTitle,
				TotalQuestions: len(h.questions),
				Participants:   count,
				Mode:           mode,
			})
			client.SafeSend(welcome)

			// Late join: if live mode has an active question, send it to the new client
			if !h.SelfPaced && h.currentQ >= 0 && h.currentQ < len(h.questions) {
				question := h.questions[h.currentQ]
				var options []string
				_ = json.Unmarshal(question.Options, &options)

				qMsg, _ := NewMessage("new_question", NewQuestionPayload{
					QuestionIdx: h.currentQ,
					Text:        question.Text,
					Options:     options,
					TimeLimit:   h.TimePerQuestion,
				})
				client.SafeSend(qMsg)

				// Also send current answer progress
				progressMsg, _ := NewMessage("answer_progress", AnswerProgressPayload{
					QuestionIdx:   h.currentQ,
					AnsweredCount: len(h.answeredUsers),
					TotalPlayers:  count,
				})
				client.SafeSend(progressMsg)
			}

			h.mu.Unlock()

			// Broadcast join to everyone
			joined, _ := NewMessage("player_joined", PlayerJoinedPayload{
				Username:         client.Username,
				ParticipantCount: count,
			})
			h.broadcastMessage(joined)

			// Send updated leaderboard to everyone on join
			h.broadcastLeaderboard()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			count := len(h.clients)
			h.mu.Unlock()

			// Broadcast updated participant count
			left, _ := NewMessage("player_left", PlayerJoinedPayload{
				Username:         client.Username,
				ParticipantCount: count,
			})
			h.broadcastMessage(left)

		case message := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				if !client.SafeSend(message) {
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

// DisconnectUser removes any existing client with the same userID (for reconnection).
func (h *Hub) DisconnectUser(userID string) {
	h.mu.Lock()
	for client := range h.clients {
		if client.UserID == userID {
			client.Close()
			delete(h.clients, client)
			go func(c *Client) { _ = c.conn.Close() }(client)
			break
		}
	}
	h.mu.Unlock()
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsUsernameTaken checks if a username is already connected in this hub.
func (h *Hub) IsUsernameTaken(username string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.clients {
		if client.Username == username {
			return true
		}
	}
	return false
}

func (h *Hub) broadcastMessage(msg []byte) {
	h.broadcast <- msg
}

func (h *Hub) HandleSubmitAnswer(client *Client, payload json.RawMessage) {
	var p SubmitAnswerPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		sendError(client, "invalid payload")
		return
	}

	h.mu.RLock()
	expectedQ := h.currentQ
	if h.SelfPaced {
		expectedQ = client.currentQ
	}
	if p.QuestionIdx < 0 || p.QuestionIdx >= len(h.questions) || p.QuestionIdx != expectedQ {
		h.mu.RUnlock()
		sendError(client, "invalid question")
		return
	}
	question := h.questions[p.QuestionIdx]
	h.mu.RUnlock()

	result, err := h.scorer.ProcessAnswer(
		context.Background(),
		h.QuizCode,
		client.UserID,
		p.QuestionIdx,
		p.SelectedIdx,
		question,
		p.TimeRemaining,
		h.TimePerQuestion,
	)
	if err != nil {
		sendError(client, err.Error())
		return
	}

	answerMsg, _ := NewMessage("answer_result", AnswerResultPayload{
		IsCorrect:     result.IsCorrect,
		PointsAwarded: result.PointsAwarded,
		CorrectIdx:    result.CorrectIdx,
		YourTotal:     result.YourTotal,
	})
	client.SafeSend(answerMsg)

	// Track answer progress for live mode
	if !h.SelfPaced {
		h.mu.Lock()
		h.answeredUsers[client.UserID] = true
		answeredCount := len(h.answeredUsers)
		totalPlayers := len(h.clients)
		h.mu.Unlock()

		progressMsg, _ := NewMessage("answer_progress", AnswerProgressPayload{
			QuestionIdx:   p.QuestionIdx,
			AnsweredCount: answeredCount,
			TotalPlayers:  totalPlayers,
		})
		h.broadcastMessage(progressMsg)
	}

	h.broadcastLeaderboard()
}

func (h *Hub) HandleNextQuestion(client *Client) {
	if h.SelfPaced {
		h.handleNextQuestionSelfPaced(client)
		return
	}

	// Only host can advance questions in live mode
	if client.UserID != h.HostUserID {
		sendError(client, "only the host can advance questions")
		return
	}

	h.mu.Lock()
	h.currentQ++
	idx := h.currentQ
	// Reset answer tracking for new question
	h.answeredUsers = make(map[string]bool)
	h.mu.Unlock()

	if idx >= len(h.questions) {
		h.broadcastQuizFinished()
		return
	}

	question := h.questions[idx]
	var options []string
	_ = json.Unmarshal(question.Options, &options)

	msg, _ := NewMessage("new_question", NewQuestionPayload{
		QuestionIdx: idx,
		Text:        question.Text,
		Options:     options,
		TimeLimit:   h.TimePerQuestion,
	})
	h.broadcastMessage(msg)
}

func (h *Hub) handleNextQuestionSelfPaced(client *Client) {
	client.currentQ++
	idx := client.currentQ

	if idx >= len(h.questions) {
		// Send finish only to this client
		entries, _ := h.leaderboard.GetLeaderboard(context.Background(), h.QuizCode, -1)
		wsEntries := make([]LeaderboardEntry, len(entries))
		for i, e := range entries {
			wsEntries[i] = LeaderboardEntry(e)
		}
		msg, _ := NewMessage("quiz_finished", LeaderboardUpdatePayload{Rankings: wsEntries})
		client.SafeSend(msg)
		return
	}

	question := h.questions[idx]
	var options []string
	_ = json.Unmarshal(question.Options, &options)

	msg, _ := NewMessage("new_question", NewQuestionPayload{
		QuestionIdx: idx,
		Text:        question.Text,
		Options:     options,
		TimeLimit:   h.TimePerQuestion,
	})
	client.SafeSend(msg)
}

func (h *Hub) broadcastLeaderboard() {
	entries, err := h.leaderboard.GetLeaderboard(context.Background(), h.QuizCode, 10)
	if err != nil {
		log.Printf("failed to get leaderboard: %v", err)
		return
	}

	wsEntries := make([]LeaderboardEntry, len(entries))
	for i, e := range entries {
		wsEntries[i] = LeaderboardEntry(e)
	}

	msg, _ := NewMessage("leaderboard_update", LeaderboardUpdatePayload{Rankings: wsEntries})
	h.broadcastMessage(msg)
}

func (h *Hub) broadcastQuizFinished() {
	entries, _ := h.leaderboard.GetLeaderboard(context.Background(), h.QuizCode, -1)

	wsEntries := make([]LeaderboardEntry, len(entries))
	for i, e := range entries {
		wsEntries[i] = LeaderboardEntry(e)
	}

	msg, _ := NewMessage("quiz_finished", LeaderboardUpdatePayload{Rankings: wsEntries})
	h.broadcastMessage(msg)
}

func sendError(client *Client, message string) {
	msg, _ := NewMessage("error", ErrorPayload{Message: message})
	client.SafeSend(msg)
}

type HubManager struct {
	hubs map[string]*Hub
	mu   sync.RWMutex
}

func NewHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[string]*Hub),
	}
}

func (hm *HubManager) GetHub(quizCode string) (*Hub, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	hub, ok := hm.hubs[quizCode]
	return hub, ok
}

func (hm *HubManager) CreateHub(hub *Hub) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.hubs[hub.QuizCode] = hub
	go hub.Run()
}

func (hm *HubManager) RemoveHub(quizCode string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	delete(hm.hubs, quizCode)
}
