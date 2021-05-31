package question

import (
	"encoding/json"
	"fmt"
	"github.com/akselleirv/introspect/models"
	"io"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	NumberOfQuestionsToFind = 4
)

type Questioner interface {
	// GetFourUnique gets a slice of question ids of the question that already have been received
	// then it returns a slice of four new questions
	GetFourUnique(usedIds []string) ([]models.Question, error)
}

type Store struct {
	q models.Questions
}

func NewStore(filePath string) *Store {
	s := Store{q: load(filePath)}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(s.q.Questions),
		func(i, j int) {
			s.q.Questions[i], s.q.Questions[j] = s.q.Questions[j], s.q.Questions[i]
		})

	return &s
}

func (s *Store) GetFourUnique(usedIds []string) ([]models.Question, error) {
	var result []models.Question
	for _, q := range s.q.Questions {
		if isNew(usedIds, q.Id) {
			result = append(result, q)
		}
		if len(result) == NumberOfQuestionsToFind {
			break
		}
	}
	if len(result) != NumberOfQuestionsToFind {
		return nil, fmt.Errorf("unable to find 4 questions, found %d", len(result))
	}
	return result, nil
}

// isNew checks if the question id is new / has been used
func isNew(usedIds []string, questionId string) bool {
	for _, id := range usedIds {
		if id == questionId {
			return false
		}
	}
	return true
}

// load loads the questions and set them in the Store struct
func load(filePath string) models.Questions {
	f, err := os.Open(filePath)
	if err != nil {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatalln("unable to get current dir: ", err)
		}
		log.Println("current dir is: ", pwd)
		log.Fatalln("unable to open 'question.json")
	}
	fr, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln("unable to read file: ", err)
	}

	var questions models.Questions
	err = json.Unmarshal(fr, &questions)
	if err != nil {
		log.Println("unable to unmarshal file: ", err)
	}
	return questions
}
