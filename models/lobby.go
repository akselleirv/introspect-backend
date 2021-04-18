package models

type LobbyUpdateAction string

const (
	Joined LobbyUpdateAction = "JOINED"
	Left                     = "LEFT"
)

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
	Event         string             `json:"event"`
	Players       []PlayerUpdate     `json:"players"`
	IsAllReady    bool               `json:"isAllReady"`
	ActionTrigger LobbyActionTrigger `json:"actionTrigger,omitempty"`
}

type LobbyActionTrigger struct {
	Player string            `json:"player"`
	Action LobbyUpdateAction `json:"action"`
}

type PlayerUpdate struct {
	Name    string `json:"name"`
	IsReady bool   `json:"isReady"`
}
