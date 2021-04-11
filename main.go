package main

import (
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

	newConn := func(w http.ResponseWriter, r *http.Request) {
		player, room := getParams(r)
		if player == "" || room == "" {
			fmt.Fprint(w, "room name or playername is missing from the URL")
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		s.NewConn(c, player, room )
	}

	http.HandleFunc("/ws", newConn)
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
		log.Println("unable to find playerName in URL")
		return "", ""
	}

	room, ok := r.URL.Query()["room"]
	if !ok || room[0] == "" {
		log.Println("unable to find room name in URL")
		return "", ""
	}

	return player[0], room[0]
}
