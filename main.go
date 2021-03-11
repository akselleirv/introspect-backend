package main

import (
	"github.com/akselleirv/introspect/events"
	"github.com/akselleirv/introspect/server"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)



var upgrader = websocket.Upgrader{}

func main() {
	s := server.NewServer()
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	events.Setup(s)

	newConn := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		s.NewConn(c, "test")
	}

	http.HandleFunc("/ws", newConn)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

