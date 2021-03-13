package models

type PlayerJoined struct {
	Event string `json:"event"`
	Name  string `json:"name"`
}
type PlayerLeft struct {
	Event string `json:"event"`
	Name  string `json:"name"`
}

type Ping struct {
	Event  string `json:"event"`
	Player string `json:"player"`
}
