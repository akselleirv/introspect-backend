package game

import (
	"fmt"
	"github.com/akselleirv/introspect/models"
	"log"
	"sync"
)

type Gamer interface {
	// set the player ready to play
	SetPlayerReady(playerName string) error

	// return true if all players are ready
	// and a slice of players who are ready
	IsPlayersReady() (bool, []string)
	AddPlayer(playerName string) bool
	RemovePlayer(playerName string)
	GetRoomStatus() ([]models.PlayerUpdate, bool)
}

type Game struct {
	players map[string]*player
	mu      sync.RWMutex
}

type player struct {
	ready bool
	// points etc..
}

func NewGame() Game {
	return Game{
		players: make(map[string]*player),
		mu:      sync.RWMutex{},
	}
}

func (g *Game) SetPlayerReady(playerName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if p, exist := g.players[playerName]; exist {
		p.ready = true
		log.Println("setting player ready: ", playerName)
		return nil
	} else {
		err := fmt.Errorf("unable to find a player with the name '%s', when setting ready status to true", playerName)
		log.Println(err)
		return err
	}
}

func (g *Game) IsPlayersReady() (bool, []string) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var readyPlayers []string

	for name, p := range g.players {
		log.Println(name, p)
		if p.ready {
			readyPlayers = append(readyPlayers, name)
		}
	}

	return len(readyPlayers) == len(g.players), readyPlayers
}

func (g *Game) AddPlayer(playerName string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, exists := g.players[playerName]; exists {
		return false
		// TODO: handle error - that name is taken
	} else {
		g.players[playerName] = &player{ready: false}
		return true
	}
}

func (g *Game) RemovePlayer(playerName string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.players, playerName)
}

// GetRoomStatus will get the status of the players
// by returning a map of clientName and a boolean if the player is ready or not
// and a boolean which is true when all players are ready
func (g *Game) GetRoomStatus() ([]models.PlayerUpdate, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var playersUpdate []models.PlayerUpdate
	isAllReady := true
	for name, player := range g.players {
		if !player.ready {
			isAllReady = false
		}
		playersUpdate = append(playersUpdate, models.PlayerUpdate{
			Name:    name,
			IsReady: player.ready,
		})
	}
	return playersUpdate, isAllReady
}
