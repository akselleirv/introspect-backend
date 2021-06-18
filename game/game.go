package game

import (
	"fmt"
	"github.com/akselleirv/introspect/models"
	"github.com/akselleirv/introspect/question"
	"log"
	"sync"
)

const (
	MaxVotesPerQuestion = 2
	QuestionsPerRound   = 4
	FirstQuestionNumber = 1
)

type SelfVote string

const (
	MostVoted  SelfVote = "Most Voted"
	Neutral    SelfVote = "Neutral"
	LeastVoted SelfVote = "Least Voted"

	MostVotedPoints  = 3
	NeutralPoints    = 1
	LeastVotedPoints = 3
	WrongVotedPoints = 0
)

type Gamer interface {
	SetPlayerReadyToStartGame(playerName string) error

	// return true if all players are readyToStartGame
	// and a slice of players who are readyToStartGame
	IsPlayersReady() (bool, []string)
	AddPlayer(playerName string) bool
	RemovePlayer(playerName string)
	GetRoomStatus() ([]models.PlayerUpdate, bool)

	// GetQuestions will return four question that
	// the room has not yet received
	GetQuestions() ([]models.Question, error)
	GetCurrentDoneQuestion() int
	SetVotesFromPlayer(question models.PlayerVotedOnQuestion)
	IsSelfVoting() bool
	SetSelfVoteFromPlayer(vote models.RegisterSelfVote)
	IsRoundFinished() (bool, bool)

	CalculatePointsForCurrentQuestion() models.QuestionPoints
	CalculatePoints(from, to int) []models.PointsEntrySimple

	SetPlayerReadyForNextRound(playerName string) error
	IsNextRound() bool

	AddCustomQuestion(question string)
}

type Game struct {
	players         map[string]*player
	currentQuestion int
	customQuestions []models.Question
	questions       []models.Question
	questionStore   question.Questioner
	mu              sync.RWMutex
}

type player struct {
	readyToStartGame  bool
	readyForNextRound bool
	// a map of question number and number of votes the player have received
	votes     map[int]int
	selfVotes map[int]SelfVote
}

func NewGame(questionFilePath string) Game {
	return Game{
		players:         make(map[string]*player),
		currentQuestion: 1,
		questionStore:   question.NewStore(questionFilePath),
		mu:              sync.RWMutex{},
	}
}

// GetLatestQuestion returns the last question number that was answered
func (g *Game) GetCurrentDoneQuestion() int {
	return g.currentQuestion - 1
}

func (g *Game) SetPlayerReadyToStartGame(playerName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if p, exist := g.players[playerName]; exist {
		p.readyToStartGame = true
		log.Println("setting player readyToStartGame: ", playerName)
		return nil
	} else {
		err := fmt.Errorf("unable to find a player with the name '%s', when setting readyToStartGame status to true", playerName)
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
		if p.readyToStartGame {
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
		g.players[playerName] = &player{readyToStartGame: false, votes: make(map[int]int), selfVotes: map[int]SelfVote{}}
		return true
	}
}

func (g *Game) RemovePlayer(playerName string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.players, playerName)
}

// GetRoomStatus will get the status of the players
// by returning a map of clientName and a boolean if the player is readyToStartGame or not
// and a boolean which is true when all players are readyToStartGame
func (g *Game) GetRoomStatus() ([]models.PlayerUpdate, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var playersUpdate []models.PlayerUpdate
	isAllReady := true
	for name, player := range g.players {
		if !player.readyToStartGame {
			isAllReady = false
		}
		playersUpdate = append(playersUpdate, models.PlayerUpdate{
			Name:    name,
			IsReady: player.readyToStartGame,
		})
	}
	return playersUpdate, isAllReady
}

// loadQuestions will make the call to the question database and set it question on the Game struct
func (g *Game) loadQuestions() error {
	result := make([]models.Question, QuestionsPerRound)
	if g.isCustomQuestions() {
		result = g.getCustomQuestions()
	}
	newQuestions, err := g.questionStore.GetFourUnique(getQuestionIds(g.questions))
	if err != nil {
		return err
	}
	result = append(result, newQuestions...)
	g.questions = append(g.questions, result[0:QuestionsPerRound]...)
	return nil
}
func (g *Game) isCustomQuestions() bool {
	return len(g.customQuestions) > 0
}

func (g *Game) getCustomQuestions() []models.Question {
	result := make([]models.Question, QuestionsPerRound)
	var i int
	for i = 0; i < len(g.customQuestions); i++ {
		result[i] = g.customQuestions[i]
	}
	g.customQuestions = g.customQuestions[i:]
	return result[:i]
}

func getQuestionIds(qs []models.Question) []string {
	var ids []string
	for _, q := range qs {
		ids = append(ids, q.Id)
	}
	return ids
}

// GetQuestions will return the last four questions
func (g *Game) GetQuestions() ([]models.Question, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.questions) <= g.currentQuestion {
		err := g.loadQuestions()
		if err != nil {
			return nil, fmt.Errorf("no more questions: %w", err)
		}
	}

	lastFourQuestions := g.questions[len(g.questions)-4:]
	return lastFourQuestions, nil
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

func (g *Game) IsSelfVoting() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var totalVotes int
	for _, p := range g.players {
		for _, votes := range p.votes {
			totalVotes += votes
		}
	}
	expectedTotalVotes := g.currentQuestion * MaxVotesPerQuestion * len(g.players)
	return expectedTotalVotes == totalVotes
}

func (g *Game) SetSelfVoteFromPlayer(vote models.RegisterSelfVote) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if p, exist := g.players[vote.Player]; exist {
		p.selfVotes[g.currentQuestion] = SelfVote(vote.Choice)
	}
}

// HaveAllPlayersSelfVoted check if all the players have issued their self vote.
// It then returns two booleans.
// First is true if all players have issued their self vote.
// Second is true if players have self voted AND it is the last round.
func (g *Game) IsRoundFinished() (bool, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	var totalSelfVotesForRound int
	for _, p := range g.players {
		if p.selfVotes[g.currentQuestion] != "" {
			totalSelfVotesForRound++
		}
	}

	questionDone := len(g.players) == totalSelfVotesForRound
	roundFinished := questionDone && g.currentQuestion%QuestionsPerRound == 0
	log.Printf("current round: '%d' - questionDone: '%t' - roundFinished: '%t' ", g.currentQuestion, questionDone, roundFinished)

	if questionDone {
		g.currentQuestion++
	}
	return questionDone, roundFinished
}

// CalculatePoints calculates points from the given range of questions
func (g *Game) CalculatePoints(from, to int) []models.PointsEntrySimple {
	g.mu.RLock()
	defer g.mu.RUnlock()
	totalPoints := make(map[string]int)
	for i := from; i <= to; i++ {
		qp := getPointsForQuestion(g.players, i)
		for _, entry := range qp {
			totalPoints[entry.Player] += entry.Points
		}
	}
	var pes []models.PointsEntrySimple
	for player, points := range totalPoints {
		pes = append(pes, models.PointsEntrySimple{
			Player: player,
			Points: points,
		})
	}
	return pes
}

func (g *Game) CalculatePointsForCurrentQuestion() models.QuestionPoints {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return getPointsForQuestion(g.players, g.currentQuestion-1)
}

type playerStat struct {
	name     string
	votes    int
	selfVote SelfVote
}

func getPointsForQuestion(players map[string]*player, currentRound int) models.QuestionPoints {
	vs := getPlayerStats(players, currentRound)
	min, max := findMinAndMaxVotes(vs)
	leastVoted, neutral, mostVoted := findPlayerPositions(vs, min, max)
	return givePoints(leastVoted, neutral, mostVoted)
}

// getPlayerStats converts the player map for the currentRound into a slice
func getPlayerStats(players map[string]*player, currentRound int) []playerStat {
	var ps []playerStat
	for n, p := range players {
		ps = append(ps, playerStat{name: n, votes: p.votes[currentRound], selfVote: p.selfVotes[currentRound]})
	}
	return ps
}

func findMinAndMaxVotes(playerStats []playerStat) (min, max int) {
	min = playerStats[0].votes
	max = playerStats[0].votes
	for _, playerStat := range playerStats {
		if playerStat.votes < min {
			min = playerStat.votes
		}
		if playerStat.votes > max {
			max = playerStat.votes
		}
	}
	return min, max
}
func findPlayerPositions(playerStats []playerStat, min, max int) (leastVoted, neutral, mostVoted []playerStat) {
	for _, playerStat := range playerStats {
		if playerStat.votes == max {
			mostVoted = append(mostVoted, playerStat)
		} else if playerStat.votes == min {
			leastVoted = append(leastVoted, playerStat)
		} else {
			neutral = append(neutral, playerStat)
		}
	}
	return leastVoted, neutral, mostVoted
}

func givePoints(leastVoted, neutral, mostVoted []playerStat) models.QuestionPoints {
	var qp models.QuestionPoints

	calculatePoints := func(s []playerStat, sv SelfVote, pointToGiveOnCorrect int) {
		for _, p := range s {
			pe := models.PointsEntry{Player: p.name, SelfVote: string(p.selfVote), VotesReceived: p.votes}
			qp = append(qp, pe)
			if p.selfVote == sv {
				qp[len(qp)-1].Points = pointToGiveOnCorrect

			} else {
				qp[len(qp)-1].Points = WrongVotedPoints

			}
		}
	}
	calculatePoints(leastVoted, LeastVoted, LeastVotedPoints)
	calculatePoints(neutral, Neutral, NeutralPoints)
	calculatePoints(mostVoted, MostVoted, MostVotedPoints)
	return qp
}

// SetNextRoundFromPlayer sets the ready for next round flag to true
func (g *Game) SetPlayerReadyForNextRound(playerName string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if p, exist := g.players[playerName]; exist {
		p.readyForNextRound = true
		return nil
	} else {
		err := fmt.Errorf("unable to find a player with the name '%s', when setting readyToStartGame status to true", playerName)
		log.Println(err)
		return err
	}
}

// IsNextRound returns true if all player have set ready for next round flag
func (g *Game) IsNextRound() bool {
	allReady := true
	g.mu.RLock()
	for _, p := range g.players {
		if p.readyForNextRound == false {
			allReady = false
		}
	}
	g.mu.RUnlock()

	if allReady {
		g.resetAllReadyForNextRound()
	}
	return allReady
}

// resetAllReadyForNextRound sets all ready for next round flags to false
// it used before starting a new round
func (g *Game) resetAllReadyForNextRound() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, p := range g.players {
		p.readyForNextRound = false
	}
}

func (g *Game) AddCustomQuestion(question string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// we do not know the language, but we still want the Question type
	g.customQuestions = append(g.customQuestions, models.Question{
		Id:       "",
		Question: models.QuestionTranslations{Norwegian: question, English: question},
	})
}
