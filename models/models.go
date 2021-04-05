package models

type Ping struct {
	Event  string `json:"event"`
	Player string `json:"player"`
}

type GenericEvent struct {
	Event  string `json:"event"`
	Player string `json:"player"`
}

type LobbyChat struct {
	Event   string `json:"event"`
	Player  string `json:"player"`
	Message string `json:"message"`
}

type LobbyRoomUpdate struct {
	Event      string         `json:"event"`
	Players    []PlayerUpdate `json:"players"`
	IsAllReady bool           `json:"isAllReady"`
}

type PlayerUpdate struct {
	Name    string `json:"name"`
	IsReady bool   `json:"isReady"`
}

type Question struct {
	QuestionID string `json:"questionID"`
	Question   string `json:"question"`
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
type QuestionPoints []PointsEntry
type TotalPoints map[int]QuestionPoints

type QuestionPointsEvent struct {
	Event          string         `json:"event"`
	QuestionPoints QuestionPoints `json:"questionPoints"`
}
