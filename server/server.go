package server

import (
	"github.com/gorilla/websocket"
	"log"
)

type Server interface {
	AddEvent(eventName string, fn eventHandler)
	Broadcast(msg string)
	SendMsg(player, msg string)
}

type Serve struct {
	EventHandlers map[string]eventHandler
	players       map[string]player
}

type player struct {
	name      string
	sendMsgCh chan<- string
}

type eventHandler = func(msg map[string]interface{})

const Event = "event"

func (s *Serve) handleMsg(msg map[string]interface{}) {
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

func (s *Serve) NewConn(c *websocket.Conn, playerName string) {
	ch := make(chan string)
	p := player{name: playerName, sendMsgCh: ch}
	s.players[playerName] = p

	go s.readMessages(c)
	go writeMessage(c, ch)
}

func NewServer() *Serve {
	return &Serve{EventHandlers: make(map[string]eventHandler), players: make(map[string]player)}
}

func (s *Serve) readMessages(c *websocket.Conn) {
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

func writeMessage(c *websocket.Conn, msgToSend <-chan string) {
	var err error
	for m := range msgToSend {
		err = c.WriteMessage(1, []byte(m))
		if err != nil {
			log.Println("unable to write to client: ", err)
		}
	}

}

func (s *Serve) AddEvent(eventName string, fn eventHandler) {
	s.EventHandlers[eventName] = fn
}

func (s *Serve) Broadcast(msg string) {

}

func (s *Serve) SendMsg(player, msg string) {
	p, ok := s.players[player]
	if !ok {
		log.Println("unable to find player: ", player)
		return
	}
	p.sendMsgCh <- msg
}
