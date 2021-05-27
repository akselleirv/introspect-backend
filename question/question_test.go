package question

import (
	"github.com/akselleirv/introspect/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	LengthUIID   = 36
	TestFilePath = "../testQuestions.json"
)

func TestLoadSuccess(t *testing.T) {
	assert.NotPanics(t, func() { load(TestFilePath) })
	q := load(TestFilePath)
	if len(q.Questions) < 2 {
		t.Fatal("test require at least two questions in questions.json")
	}
	assert.Equal(t, LengthUIID, len(q.Questions[0].Id), "question id length is of wrong length")
	assert.Equal(t, LengthUIID, len(q.Questions[1].Id), "question id length is of wrong length")
}

func TestStore_GetFourUniqueSuccess(t *testing.T) {
	s := NewStore(TestFilePath)
	q1s, err := s.GetFourUnique([]string{})

	assert.NoError(t, err)
	assert.Len(t, q1s, NumberOfQuestionsToFind)

	q2s, err := s.GetFourUnique(getQuestionIds(q1s))
	assert.NoError(t, err)
	assert.Len(t, q2s, NumberOfQuestionsToFind)

	for _, q1 := range q1s {
		for _, q2 := range q2s {
			assert.NotEqual(t, q1, q2, "all the questions should be unique")
		}
	}
}

func TestStore_GetFourUniqueNoMoreQuestions(t *testing.T) {
	s := NewStore(TestFilePath)
	var ids []string
	var err error
	var qs []models.Question

	for i := 0; i < 2; i++ {
		qs, err = s.GetFourUnique(ids)
		assert.NoError(t, err)
		ids = append(ids, getQuestionIds(qs)...)
	}

	// now we have requested all the question in the testQuestions.json
	assert.Len(t, ids, 2*NumberOfQuestionsToFind)
	qs, err = s.GetFourUnique(ids)
	assert.Nil(t, qs)
	assert.EqualError(t, err, "unable to find 4 questions, found 0", "we requested the same ids we should got, it should result in an error")
}

func getQuestionIds(qs []models.Question) []string {
	var ids []string
	for _, q := range qs {
		ids = append(ids, q.Id)
	}
	return ids
}
