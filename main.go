package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type eventHandler = func(msg map[string]interface{})

const Event = "event"

type Server struct {
	EventHandlers map[string]eventHandler
}

type Message struct {
	Player string `json:"player"`
}

var upgrader = websocket.Upgrader{}

func (s *Server) handleMsg(msg map[string]interface{}) {
	e, ok := msg[Event].(string)
	if !ok {
		log.Println("sent event is not a string")
		return
	}
	delete(msg, Event)
	handler, ok := s.EventHandlers[e]
	if !ok {
		log.Println("unable to find event in event handlers: ", e)
		return
	}
	handler(msg)
}

func (s *Server) processData(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	msg := make(map[string]interface{})
	for {
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", msg)
		s.handleMsg(msg)
	}
}

func main() {
	s := NewServer()
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	s.addEvent("join", func(data map[string]interface{}) {
		var msg Message
		parseToJson(&data, &msg)
		fmt.Println(msg.Player)

	})

	http.HandleFunc("/ws", s.processData)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func NewServer() Server {
	return Server{EventHandlers: make(map[string]eventHandler)}
}

func (s *Server) addEvent(eventName string, fn eventHandler) {
	s.EventHandlers[eventName] = fn
}

func parseToJson(data *map[string]interface{}, msg interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("unable to marshal '%s': %s", *data, err)
		return
	}
	err = json.Unmarshal(b, msg)
	if err != nil {
		log.Printf("unable to unmarshal message: %s", err.Error())
		return
	}
}
