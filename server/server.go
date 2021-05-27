package server

import (
	"fmt"
	"github.com/akselleirv/introspect/events"
	"github.com/akselleirv/introspect/handler"
	"github.com/akselleirv/introspect/models"
	"github.com/akselleirv/introspect/room"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"sync"
)

type Server interface {
	NewConn(c *websocket.Conn, playerName, roomName string)
	IsGameInfoValid(roomName, playerName string) bool
}

type Serve struct {
	rooms map[string]room.Roomer
	mu    sync.RWMutex
}

func NewServer() *Serve {
	return &Serve{rooms: make(map[string]room.Roomer), mu: sync.RWMutex{}}
}

func (s *Serve) NewConn(c *websocket.Conn, playerName, roomName string) {
	if exist := s.roomExist(roomName); !exist {
		h := handler.NewHandler(log.New(os.Stdout, "", 64))
		initEventHandlers := events.Setup(h)
		msgHandler := h.HandleMsg()
		r, err := s.createRoom(roomName, initEventHandlers, msgHandler)
		if err != nil {
			log.Println(err)
			return
		}
		s.registerNewRoom(roomName, r)
	}

	s.addPlayerToRoom(c, playerName, roomName)
}

func (s *Serve) IsGameInfoValid(roomName, playerName string) (playerNameAvailable bool, roomIsJoinable bool) {
	if r, ok := s.getRoom(roomName); ok {
		playerNameAvailable = r.IsPlayerNameAvailable(playerName)
		roomIsJoinable = r.IsRoomJoinable()
		return playerNameAvailable, roomIsJoinable
	} else {
		// room does not exist - both are then valid
		playerNameAvailable = true
		roomIsJoinable = true
		return playerNameAvailable, roomIsJoinable
	}
}

func (s *Serve) deleteRoom(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rooms[name]; ok {
		delete(s.rooms, name)
	}
}

func (s *Serve) addPlayerToRoom(c *websocket.Conn, playerName, roomName string) {
	if r, ok := s.getRoom(roomName); ok {
		r.AddClient(c, playerName)
	} else {
		log.Println("unable to find room - this error message should not printed - this is awkward")
		err := c.WriteJSON(models.GenericEvent{Event: "unable_to_find_room", Player: playerName})
		if err != nil {
			log.Printf("error sending message to player '%s': %s", playerName, err.Error())
		}
	}
}

// RoomExist returns true if the room name exist and false if it does not
func (s *Serve) roomExist(roomName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.rooms[roomName]
	return ok
}

func (s *Serve) createRoom(name string, initEventHandlers func(r room.Roomer), msgHandler func(msg map[string]interface{})) (room.Roomer, error) {
	if exist := s.roomExist(name); exist {
		return nil, fmt.Errorf("room '%s' already exists", name)
	}
	return room.NewRoom(name, initEventHandlers, msgHandler, func() { s.deleteRoom(name) }), nil
}

func (s *Serve) registerNewRoom(name string, r room.Roomer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rooms[name] = r
}

func (s *Serve) getRoom(roomName string) (room.Roomer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rooms[roomName]
	return r, ok
}
