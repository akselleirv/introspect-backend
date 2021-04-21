package main

import (
	"encoding/json"
	"fmt"
	"github.com/akselleirv/introspect/server"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{}

func main() {
	s := server.NewServer()

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		playerName, room := getParams(r)
		if playerName == "" || room == "" {
			fmt.Fprint(w, "room name or playername is missing from the URL")
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		s.NewConn(c, playerName, room)
	})

	http.HandleFunc("/validateGameInfo", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		playerName, room := getParams(req)
		if playerName == "" || room == "" {
			fmt.Fprint(w, "room name or playername is missing from the URL")
			return
		}

		p, r := s.IsGameInfoValid(room, playerName)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			PlayerNameAvailable bool `json:"playerNameAvailable"`
			RoomIsJoinable      bool `json:"roomIsJoinable"`
		}{
			PlayerNameAvailable: p,
			RoomIsJoinable:      r,
		})
		return
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "pong")
	})

	log.Println("starting server - listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// getParams returns playerName and roomName from the URL param
func getParams(r *http.Request) (string, string) {
	player, ok := r.URL.Query()["player"]
	if !ok || player[0] == "" {
		log.Println("unable to find player in URL")
		return "", ""
	}

	room, ok := r.URL.Query()["room"]
	if !ok || room[0] == "" {
		log.Println("unable to find room name in URL")
		return "", ""
	}

	return player[0], room[0]
}
