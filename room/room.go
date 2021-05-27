package room

import (
	"encoding/json"
	"github.com/akselleirv/introspect/client"
	"github.com/akselleirv/introspect/game"
	"github.com/akselleirv/introspect/models"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Roomer interface {
	AddClient(c *websocket.Conn, name string)
	Broadcast(msg []byte)
	SendMsg(clientName string, msg []byte)
	Game() game.Gamer
	IsPlayerNameAvailable(name string) bool
	IsRoomJoinable() bool
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

func (r *Room) removeClient(name string) {
	r.mu.Lock()
	delete(r.clients, name)
	r.game.RemovePlayer(name)
	log.Printf("removed client '%s' from Room '%s'", name, r.name)
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
		ActionTrigger: models.LobbyActionTrigger{
			Player: name,
			Action: models.Left,
		},
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
			ActionTrigger: models.LobbyActionTrigger{
				Player: name,
				Action: models.Joined,
			},
		})
		r.Broadcast(b)
	} else {
		log.Printf("awkward this error should never been diplayed - silently failing adding client to room - player '%s' already exist", name)
	}
}

func (r *Room) Broadcast(msg []byte) {
	for _, p := range r.clients {
		p.Send(msg)
	}
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

func (r *Room) IsPlayerNameAvailable(name string) bool {
	if _, exist := r.clients[name]; exist {
		return false
	}
	return true
}

func (r *Room) IsRoomJoinable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rdy, _ := r.Game().IsPlayersReady()
	return rdy == false
}
