package room

import (
	"encoding/json"
	"github.com/akselleirv/introspect/client"
	"github.com/akselleirv/introspect/models"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Room interface {
	AddClient(c *websocket.Conn, playerName string)
	Broadcast(msg []byte)
	SendMsg(player string, msg []byte)
}

type room struct {
	name       string
	clients    map[string]client.Clienter
	msgHandler func(msg map[string]interface{})
	deleteRoom func()
	mu         sync.RWMutex
}

func NewRoom(name string, initEventHandlers func(r Room), handleMsg func(msg map[string]interface{}), deleteRoom func()) *room {
	log.Printf("creating new room: %s", name)
	r := &room{
		name:       name,
		clients:    make(map[string]client.Clienter),
		msgHandler: handleMsg,
		deleteRoom: deleteRoom,
		mu:         sync.RWMutex{},
	}
	initEventHandlers(r)
	return r
}

func (r *room) removeClient(clientName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, clientName)
	log.Printf("removed client '%s' from room '%s'", clientName, r.name)

	if len(r.clients) == 0 {
		log.Printf("deleting room '%s' -  no more players", r.name)
		r.deleteRoom()
	}
	b, err := json.Marshal(models.PlayerLeft{
		Event: "player_left",
		Name:  clientName,
	})
	if err != nil {
		log.Println("unable to unmarshal struct when broadcasting new player")
	}

	r.Broadcast(b)
}

func (r *room) AddClient(c *websocket.Conn, playerName string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clients[playerName]; !ok {
		log.Printf("adding player '%s' to room: '%s'", playerName, r.name)
		r.clients[playerName] = client.NewClient(playerName, c, r.msgHandler, func() { r.removeClient(playerName) })
		b, err := json.Marshal(models.PlayerJoined{
			Event: "player_joined",
			Name:  playerName,
		})
		if err != nil {
			log.Println("unable to unmarshal struct when broadcasting new player")
		}
		r.Broadcast(b)
	} else {
		// playerName already exists
		// handle error
	}
}

func (r *room) Broadcast(msg []byte) {
	for _, p := range r.clients {
		p.Send(msg)
	}
}

func (r *room) SendMsg(player string, msg []byte) {
	p, ok := r.clients[player]
	if !ok {
		log.Println("unable to find player: ", player)
		return
	}
	p.Send(msg)
}
