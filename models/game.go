package models

type Questions struct {
	Questions []Question `json:"questions"`
}

type Question struct {
	Id string `json:"id"`
	// Question value is the question translated based on the key.
	// Currently supported languages are: 'no' and 'en'
	// {"questions": {"en": "question", "no": "spørsmål"}}
	Question QuestionTranslations `json:"question"`
}
type QuestionTranslations struct {
	Norwegian string `json:"no"`
	English   string `json:"en"`
}

type Vote struct {
	PlayerWhoReceivedTheVote string `json:"playerWhoReceivedTheVote"`
	QuestionID               string `json:"questionID"`
}

type PlayerVotedOnQuestion struct {
	Player string `json:"player"`
	Votes  []Vote `json:"votes"`
}

type RegisterSelfVote struct {
	Player   string   `json:"player"`
	Choice   string   `json:"choice"`
	Question Question `json:"question"`
}

type PointsEntry struct {
	Player        string `json:"player"`
	SelfVote      string `json:"selfVote"`
	VotesReceived int    `json:"votesReceived"`
	Points        int    `json:"points"`
}

type PointsEntrySimple struct {
	Player string `json:"player"`
	Points int    `json:"points"`
}
type PlayersResults struct {
	Event                        string              `json:"event"`
	PlayersResultExceptLastRound []PointsEntrySimple `json:"playersResultExceptLastRound"`
	PlayersResults               []PointsEntrySimple `json:"playersResults"`
}

type QuestionPoints []PointsEntry
type TotalPoints map[int]QuestionPoints

type QuestionPointsEvent struct {
	Event           string         `json:"event"`
	QuestionPoints  QuestionPoints `json:"questionPoints"`
	CurrentQuestion int            `json:"currentQuestion"`
}

type ErrorMsg struct {
	Event string `json:"event"`
	Error string `json:"error"`
}
