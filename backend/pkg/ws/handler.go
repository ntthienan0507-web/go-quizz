package ws

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chungnguyen/quizz-backend/modules/auth"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var guestNameRegex = regexp.MustCompile(`[^a-zA-Z0-9_ -]`)

func sanitizeGuestName(name string) string {
	return strings.TrimSpace(guestNameRegex.ReplaceAllString(name, ""))
}

const authTimeout = 5 * time.Second

type WSHandler struct {
	hubManager     *HubManager
	authService    *auth.Service
	leaderboard    LeaderboardProvider
	allowedOrigins map[string]bool
	upgrader       websocket.Upgrader
}

func NewWSHandler(hubManager *HubManager, authService *auth.Service, leaderboard LeaderboardProvider, corsOrigins string) *WSHandler {
	origins := make(map[string]bool)
	for _, o := range strings.Split(corsOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins[o] = true
		}
	}

	h := &WSHandler{
		hubManager:     hubManager,
		authService:    authService,
		leaderboard:    leaderboard,
		allowedOrigins: origins,
	}
	h.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     h.checkOrigin,
	}
	return h
}

func (h *WSHandler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return h.allowedOrigins[origin]
}

type authPayload struct {
	Token string `json:"token"`
	Guest string `json:"guest"`
}

func (h *WSHandler) Handle(c *gin.Context) {
	code := c.Param("code")
	hub, ok := h.hubManager.GetHub(code)
	if !ok {
		c.JSON(http.StatusNotFound, map[string]any{
			"status":  http.StatusNotFound,
			"message": "Quiz not found or not started",
		})
		return
	}

	// Upgrade first — no auth in query params
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Wait for auth message within timeout
	_ = conn.SetReadDeadline(time.Now().Add(authTimeout))
	_, data, err := conn.ReadMessage()
	if err != nil {
		_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "auth timeout"}})
		_ = conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{}) // clear deadline

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil || msg.Type != "auth" {
		_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "first message must be auth"}})
		_ = conn.Close()
		return
	}

	var ap authPayload
	if err := json.Unmarshal(msg.Payload, &ap); err != nil {
		_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "invalid auth payload"}})
		_ = conn.Close()
		return
	}

	var userID, username string

	if ap.Token != "" {
		claims, err := h.authService.ValidateToken(ap.Token)
		if err != nil {
			_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "invalid token"}})
			_ = conn.Close()
			return
		}
		userID = claims.UserID
		username = claims.Username
	} else if ap.Guest != "" {
		guest := sanitizeGuestName(ap.Guest)
		if guest == "" || len(guest) > 20 {
			_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "guest name must be 1-20 alphanumeric characters"}})
			_ = conn.Close()
			return
		}
		username = guest
		if !hub.SelfPaced && hub.IsUsernameTaken(username) {
			_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "username already taken in this quiz"}})
			_ = conn.Close()
			return
		}
		suffix := make([]byte, 4)
		_, _ = rand.Read(suffix)
		userID = "guest_" + hex.EncodeToString(suffix)
	} else {
		_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]string{"message": "token or guest name required"}})
		_ = conn.Close()
		return
	}

	hub.DisconnectUser(userID)

	ctx := context.Background()
	_ = h.leaderboard.SetUsername(ctx, code, userID, username)

	client := NewClient(hub, conn, userID, username)
	hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
