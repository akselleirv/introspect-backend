package room

import (
	"encoding/json"
	"github.com/akselleirv/introspect/client"
	"github.com/akselleirv/introspect/game"
	"github.com/akselleirv/introspect/models"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Roomer interface {
	AddClient(c *websocket.Conn, name string)
	Broadcast(msg []byte)
	SendMsg(clientName string, msg []byte)
	Game() game.Gamer
}

type Room struct {
	name       string
	clients    map[string]client.Clienter
	msgHandler func(msg map[string]interface{})
	deleteRoom func()
	mu         sync.RWMutex

	game game.Game
}

func NewRoom(name string, initEventHandlers func(r Roomer), handleMsg func(msg map[string]interface{}), deleteRoom func()) *Room {
	log.Printf("creating new Room: %s", name)
	r := &Room{
		name:       name,
		clients:    make(map[string]client.Clienter),
		game:       game.NewGame(),
		msgHandler: handleMsg,
		deleteRoom: deleteRoom,
		mu:         sync.RWMutex{},
	}
	initEventHandlers(r)
	return r
}

func (r *Room) removeClient(clientName string) {
	r.mu.Lock()
	delete(r.clients, clientName)
	r.game.RemovePlayer(clientName)
	log.Printf("removed client '%s' from Room '%s'", clientName, r.name)
	if len(r.clients) == 0 {
		log.Printf("deleting Room '%s' -  no more players", r.name)
		r.deleteRoom()
	}
	r.mu.Unlock()
	playersUpdate, isAllReady := r.Game().GetRoomStatus()
	b, _ := json.Marshal(models.LobbyRoomUpdate{
		Event:      "lobby_room_update",
		Players:    playersUpdate,
		IsAllReady: isAllReady,
	})
	r.Broadcast(b)
}

func (r *Room) AddClient(c *websocket.Conn, name string) {
	if _, ok := r.clients[name]; !ok {
		log.Printf("adding player '%s' to Room: '%s'", name, r.name)

		r.mu.Lock()
		r.clients[name] = client.NewClient(name, c, r.msgHandler, func() { r.removeClient(name) })
		r.mu.Unlock()

		r.game.AddPlayer(name)

		playersUpdate, isAllReady := r.Game().GetRoomStatus()
		b, _ := json.Marshal(models.LobbyRoomUpdate{
			Event:      "lobby_room_update",
			Players:    playersUpdate,
			IsAllReady: isAllReady,
		})
		r.Broadcast(b)
	} else {
		// name already exists
		// handle error
	}
}

func (r *Room) Broadcast(msg []byte) {
	t := time.Now()
	for _, p := range r.clients {
		p.Send(msg)
	}
	log.Println("Time used to broadcast: ", time.Since(t))
}

func (r *Room) SendMsg(clientName string, msg []byte) {
	p, ok := r.clients[clientName]
	if !ok {
		log.Println("unable to find clientName: ", clientName)
		return
	}
	p.Send(msg)
}

func (r *Room) Game() game.Gamer { return &r.game }
