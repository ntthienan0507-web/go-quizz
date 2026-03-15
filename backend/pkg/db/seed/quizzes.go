package seed

import (
	"encoding/json"
	"log"

	"github.com/chungnguyen/quizz-backend/modules/auth"
	"github.com/chungnguyen/quizz-backend/modules/quiz"
	"github.com/chungnguyen/quizz-backend/modules/quiz/question"
	"gorm.io/gorm"
)

func seedQuizzes(db *gorm.DB) {
	var count int64
	db.Model(&quiz.Quiz{}).Count(&count)
	if count > 0 {
		return
	}

	var admin auth.User
	if err := db.Where("email = ?", "admin@test.com").First(&admin).Error; err != nil {
		log.Printf("seed: admin user not found, skipping quiz seed")
		return
	}

	opts := func(options ...string) json.RawMessage {
		b, _ := json.Marshal(options)
		return b
	}

	quizzes := []struct {
		Quiz      quiz.Quiz
		Questions []question.Question
	}{
		{
			Quiz: quiz.Quiz{
				Title:           "General Knowledge (Live)",
				QuizCode:        "GENKNOW",
				CreatedBy:       admin.ID,
				Status:          quiz.StatusDraft,
				Mode:            quiz.ModeLive,
				TimePerQuestion: 30,
			},
			Questions: []question.Question{
				{Text: "What is the capital of France?", Options: opts("London", "Berlin", "Paris", "Madrid"), CorrectIdx: 2, Points: 10, OrderNum: 0},
				{Text: "Which planet is known as the Red Planet?", Options: opts("Venus", "Mars", "Jupiter", "Saturn"), CorrectIdx: 1, Points: 10, OrderNum: 1},
				{Text: "What is the largest ocean on Earth?", Options: opts("Atlantic", "Indian", "Arctic", "Pacific"), CorrectIdx: 3, Points: 10, OrderNum: 2},
				{Text: "Who painted the Mona Lisa?", Options: opts("Van Gogh", "Picasso", "Da Vinci", "Rembrandt"), CorrectIdx: 2, Points: 10, OrderNum: 3},
				{Text: "How many continents are there?", Options: opts("5", "6", "7", "8"), CorrectIdx: 2, Points: 10, OrderNum: 4},
			},
		},
		{
			Quiz: quiz.Quiz{
				Title:           "Science & Tech (Self-paced)",
				QuizCode:        "SCITECH",
				CreatedBy:       admin.ID,
				Status:          quiz.StatusDraft,
				Mode:            quiz.ModeSelfPaced,
				TimePerQuestion: 20,
			},
			Questions: []question.Question{
				{Text: "What does HTML stand for?", Options: opts("Hyper Text Markup Language", "High Tech Modern Language", "Hyper Transfer Markup Language", "Home Tool Markup Language"), CorrectIdx: 0, Points: 10, OrderNum: 0},
				{Text: "What is the chemical symbol for water?", Options: opts("O2", "CO2", "H2O", "NaCl"), CorrectIdx: 2, Points: 10, OrderNum: 1},
				{Text: "What is the speed of light (approx)?", Options: opts("300,000 km/s", "150,000 km/s", "500,000 km/s", "1,000,000 km/s"), CorrectIdx: 0, Points: 15, OrderNum: 2},
				{Text: "Who is known as the father of computers?", Options: opts("Alan Turing", "Charles Babbage", "Bill Gates", "Steve Jobs"), CorrectIdx: 1, Points: 10, OrderNum: 3},
				{Text: "What programming language is known as the language of the web?", Options: opts("Python", "Java", "JavaScript", "C++"), CorrectIdx: 2, Points: 10, OrderNum: 4},
			},
		},
	}

	for i := range quizzes {
		q := &quizzes[i]
		if err := db.Create(&q.Quiz).Error; err != nil {
			log.Printf("seed: failed to create quiz %s: %v", q.Quiz.Title, err)
			continue
		}

		for i := range q.Questions {
			q.Questions[i].QuizID = q.Quiz.ID
		}

		if err := db.Create(&q.Questions).Error; err != nil {
			log.Printf("seed: failed to create questions for %s: %v", q.Quiz.Title, err)
			continue
		}

		log.Printf("seed: created quiz '%s' (code: %s) with %d questions", q.Quiz.Title, q.Quiz.QuizCode, len(q.Questions))
	}
}
