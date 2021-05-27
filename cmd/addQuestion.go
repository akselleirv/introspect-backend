package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/akselleirv/introspect/models"
	"github.com/google/uuid"
	"io"
	"log"
	"os"
)

const (
	FilePath = "./questions.json"
)

func main() {
	qEn := flag.String("qEn", "", "the question to add in english")
	qNo := flag.String("qNo", "", "the question to add in norwegian")
	flag.Parse()
	if *qEn == "" && *qNo == "" {
		fmt.Println("at least one question must be added")
		os.Exit(1)
	}

	addQuestion(*qEn, *qNo)
}

func addQuestion(qEn, qNo string) {
	questions := loadQuestions()

	newQuestion := models.Question{
		Id:       newUUID(questions),
		Question: models.QuestionTranslations{Norwegian: qNo, English: qEn},
	}

	questions.Questions = append(questions.Questions, newQuestion)

	writeQuestions(questions)
}

func loadQuestions() models.Questions {
	f, err := os.Open(FilePath)
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

func writeQuestions(q models.Questions) {
	f, err := json.Marshal(q)
	if err != nil {
		log.Fatalln("unable to marshal questions: ", err)
	}
	err = os.WriteFile(FilePath, f, 0666)
	if err != nil {
		log.Fatalln("unable to write questions to file: ", err)
	}
}

func newUUID(questions models.Questions) string {
	var id string
	var i int
	for {
		id = uuid.NewString()
		if isUnique(questions, id) {
			return id
		}
		i++
		// just in case we won the lottery
		if i > 100_00 {
			log.Fatalln("holy fuck we looped a 100 000 times looking for a UUID")
		}
	}
}
func isUnique(questions models.Questions, id string) bool {
	for _, q := range questions.Questions {
		if q.Id == id {
			return false
		}
	}
	return true
}
