package game

import (
	"fmt"
	"github.com/akselleirv/introspect/models"
	"log"
	"sync"
)

const (
	MaxVotesPerQuestion = 2
	QuestionsPerRound   = 4
)

type SelfVote string

const (
	MostVoted  SelfVote = "Most Voted"
	Neutral    SelfVote = "Neutral"
	LeastVoted SelfVote = "Least Voted"
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

	// GetQuestions will return four question that
	// the room has not yet received
	GetQuestions() []models.Question
	SetVotesFromPlayer(question models.PlayerVotedOnQuestion)
	IsNextQuestionOrSelfVote() (nextQuestion bool, selfVote bool)
	RegisterSelfVote(vote models.RegisterSelfVote)
	HasAllPlayersSelfVoted() bool
}

type Game struct {
	players         map[string]*player
	currentQuestion int
	currentSelfVote int
	questions       []models.Question
	mu              sync.RWMutex
}

type player struct {
	ready bool
	// a map of question number and number of votes the player have received
	votes     map[int]int
	selfVotes map[int]SelfVote
}

func NewGame() Game {
	return Game{
		players:         make(map[string]*player),
		currentQuestion: 1,
		currentSelfVote: 1,
		mu:              sync.RWMutex{},
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
		g.players[playerName] = &player{ready: false, votes: make(map[int]int), selfVotes: map[int]SelfVote{}}
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

// loadQuestions will make the call to the question database and set it question on the Game struct
func (g *Game) loadQuestions() {
	var mockQuestion models.Question
	for i := 0; i < 4; i++ {
		mockQuestion.QuestionID = fmt.Sprintf("ID_%d", i+1) //TODO: add proper ID
		mockQuestion.Question = fmt.Sprintf("mock question number %d", i+1)
		g.questions = append(g.questions, mockQuestion)
	}
}

// GetQuestions will return the last four questions
func (g *Game) GetQuestions() []models.Question {
	g.mu.Lock()
	if len(g.questions) < g.currentQuestion {
		g.loadQuestions()
	}
	g.mu.Unlock()

	lastFourQuestions := g.questions[len(g.questions)-4:]
	return lastFourQuestions
}

// SetVotesFromPlayer register the vote from the player
func (g *Game) SetVotesFromPlayer(votes models.PlayerVotedOnQuestion) {
	g.mu.Lock()
	v1, v2 := votes.Votes[0], votes.Votes[1]
	if p, exist := g.players[v1.PlayerWhoReceivedTheVote]; exist {
		p.votes[g.currentQuestion]++
	}
	if p, exist := g.players[v2.PlayerWhoReceivedTheVote]; exist {
		p.votes[g.currentQuestion]++
	}
	g.mu.Unlock()
}

func (g *Game) IsNextQuestionOrSelfVote() (nextQuestion bool, selfVote bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.hasAllPlayersVoted() {
		if g.currentQuestion%QuestionsPerRound == 0 {
			log.Println("time for self votes!")
			selfVote = true
		} else {
			g.currentQuestion++
			nextQuestion = true
		}
	}
	return nextQuestion, selfVote
}

func (g *Game) hasAllPlayersVoted() bool {
	var totalVotes int
	for _, p := range g.players {
		for _, votes := range p.votes {
			totalVotes += votes
		}
	}
	expectedTotalVotes := g.currentQuestion * MaxVotesPerQuestion * len(g.players)
	return expectedTotalVotes == totalVotes
}

func (g *Game) RegisterSelfVote(vote models.RegisterSelfVote) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if p, exist := g.players[vote.Player]; exist {
		p.selfVotes[g.currentSelfVote] = SelfVote(vote.Choice)
	}
}

// TODO: make one generic func for counting votes
func (g *Game) HasAllPlayersSelfVoted() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var totalSelfVotes int
	for _, p := range g.players {
		if p.selfVotes[g.currentSelfVote] != "" {
			totalSelfVotes++
		}
	}
	expectedTotalSelfVotes := g.currentSelfVote * len(g.players)
	log.Printf("expected '%d', got '%d'", expectedTotalSelfVotes, totalSelfVotes)
	return expectedTotalSelfVotes == totalSelfVotes
}
