package game

import (
	"fmt"
	"github.com/akselleirv/introspect/models"
	"testing"
)

const p1, p2, p3 = "Player AAA", "Player BBB", "Player CCC"
const p1Votes, p2Votes, p3Votes = 4, 2, 0

func TestAddPlayer(t *testing.T) {
	g := NewGame()
	players := []string{"Player AAA", "Player BBB"}
	var ok bool
	ok = g.AddPlayer(players[0])
	ok = g.AddPlayer(players[1])
	if !ok {
		t.Error("unable to add player")
	}
	for name := range g.players {
		if name != players[0] && name != players[1] {
			t.Errorf("expected name to be either '%s' or '%s', got '%s'", players[0], players[1], name)
		}
	}
	ok = g.AddPlayer(players[0])
	if ok {
		t.Errorf("name '%s' already exists, expected false return value when adding player", players[0])
	}

}

func TestCalculatePointsForAllRounds(t *testing.T) {
	g := createTestableGame(t)

	// we want to finish all the rounds -  there are 3 lefts
	for i := 1; i < QuestionsPerRound; i++ {
		g.SetVotesFromPlayer(createTwoVotes(p1, p2))
		g.SetVotesFromPlayer(createTwoVotes(p2, p1))
		g.SetVotesFromPlayer(createTwoVotes(p3, p1))
		g.SetSelfVoteFromPlayer(createSelfVote(p1, MostVoted))
		g.SetSelfVoteFromPlayer(createSelfVote(p2, Neutral))
		g.SetSelfVoteFromPlayer(createSelfVote(p3, LeastVoted))

		roundFinished, _ := g.IsRoundFinished()
		if !roundFinished {
			t.Errorf("expected round to be finished")
		}
	}
	// we add +1 since the round is done and thus making the "currentQuestion" to be 5
	if g.currentQuestion != QuestionsPerRound+1 {
		t.Errorf("expected current round to be %d, got %d", QuestionsPerRound, g.currentQuestion)
	}
	totalPointsAllRounds := g.CalculatePoints(1, 4)
	for _, entry := range totalPointsAllRounds {
		if (entry.Player == p1 || entry.Player == p3) && entry.Points != MostVotedPoints*QuestionsPerRound {
			t.Errorf("expected player '%s' to have '%d' points, got '%d'", entry.Player, MostVotedPoints*QuestionsPerRound, entry.Points)
		}
		if entry.Player == p2 && entry.Points != NeutralPoints*QuestionsPerRound {
			t.Errorf("expected player '%s' to have '%d' points, got '%d'", entry.Player, NeutralPoints*QuestionsPerRound, entry.Points)
		}
	}
	fmt.Println(totalPointsAllRounds)
}

func TestFindMinAndMaxVotes(t *testing.T) {
	const Max, Min = 2, 1
	players := []string{"Player AAA", "Player BBB"}
	ps := []playerStat{
		{players[0], Max, "not_used_here"},
		{players[1], Min, "not_used_here"},
	}
	min, max := findMinAndMaxVotes(ps)
	if max != Max {
		t.Errorf("expected max to be '%d', got '%d' ", Max, max)
	}
	if min != Min {
		t.Errorf("expected min to be '%d', got '%d' ", Min, min)
	}
}

func TestGetVoteStats(t *testing.T) {
	g := createTestableGame(t)
	ps := getPlayerStats(g.players, 1)
	for _, p := range ps {
		switch p.name {
		case p1:
			if p.selfVote != MostVoted {
				t.Errorf("expected self vote '%s', got '%s' ", MostVoted, p.selfVote)
			}
			if p.votes != p1Votes {
				t.Errorf("expectes votes '%d', got '%d'", p1Votes, p.votes)
			}
		case p2:
			if p.selfVote != Neutral {
				t.Errorf("expected self vote '%s', got '%s' ", Neutral, p.selfVote)
			}
			if p.votes != p2Votes {
				t.Errorf("expectes votes '%d', got '%d'", p2Votes, p.votes)
			}
		case p3:
			if p.selfVote != LeastVoted {
				t.Errorf("expected self vote '%s', got '%s' ", LeastVoted, p.selfVote)
			}
			if p.votes != p3Votes {
				t.Errorf("expectes votes '%d', got '%d'", p3Votes, p.votes)
			}
		default:
			t.Errorf("unexpected name")
		}

	}
}

func TestFindPlayerPositions(t *testing.T) {
	g := createTestableGame(t)
	lv, n, mv := findPlayerPositions(getPlayerStats(g.players, 1), p3Votes, p1Votes)
	if len(lv) != 1 || len(n) != 1 || len(mv) != 1 {
		t.Errorf("expected all slices to be 1, got lv => '%d', n => '%d', mv => '%d'", len(lv), len(n), len(mv))
	}
	if lv[0].name != p3 || lv[0].selfVote != LeastVoted || lv[0].votes != p3Votes {
		t.Errorf("expected name '%s' to be '%s' and have '%d' votes, got name '%s', '%d' points and '%s'", p3, LeastVoted, p3Votes, lv[0].name, lv[0].votes, lv[0].selfVote)
	}
	if n[0].name != p2 || n[0].selfVote != Neutral || n[0].votes != p2Votes {
		t.Errorf("expected name '%s' to be '%s' and have '%d' votes, got name '%s', '%d' points and '%s'", p2, Neutral, p2Votes, n[0].name, n[0].votes, n[0].selfVote)
	}
	if mv[0].name != p1 || mv[0].selfVote != MostVoted || mv[0].votes != p1Votes {
		t.Errorf("expected name '%s' to be '%s' and have '%d' votes, got name '%s', '%d' points and '%s'", p1, MostVoted, p1Votes, mv[0].name, mv[0].votes, mv[0].selfVote)
	}
}

func TestGivePoints(t *testing.T) {
	const p4 = "Player DDD"
	const p4Votes = 2
	g := createTestableGame(t)
	lv, n, mv := findPlayerPositions(getPlayerStats(g.players, 1), p3Votes, p1Votes)
	// adding player who should receive zero points
	n = append(n, playerStat{
		name:     p4,
		votes:    p4Votes,
		selfVote: MostVoted,
	})
	qp := givePoints(lv, n, mv)
	for _, point := range qp {
		switch point.Player {
		case p1:
			if point.Points != MostVotedPoints {
				t.Errorf("expected player to receive '%d' points, got '%d'", MostVotedPoints, point.Points)
			}
		case p2:
			if point.Points != NeutralPoints {
				t.Errorf("expected player to receive '%d' points, got '%d'", NeutralPoints, point.Points)
			}
		case p3:
			if point.Points != LeastVotedPoints {
				t.Errorf("expected player to receive '%d' points, got '%d'", LeastVotedPoints, point.Points)
			}
		case p4:
			if point.Points != WrongVotedPoints {
				t.Errorf("expected player to receive '%d' points, got '%d'", WrongVotedPoints, point.Points)
			}
		default:
			t.Errorf("unexpected case, got '%s'", point.Player)
		}

	}

}

func createTestableGame(t *testing.T) *Game {
	g := NewGame()
	g.AddPlayer(p1)
	g.AddPlayer(p2)
	g.AddPlayer(p3)
	g.SetVotesFromPlayer(createTwoVotes(p1, p2))
	g.SetVotesFromPlayer(createTwoVotes(p2, p1))
	g.SetVotesFromPlayer(createTwoVotes(p3, p1))
	g.SetSelfVoteFromPlayer(createSelfVote(p1, MostVoted))
	g.SetSelfVoteFromPlayer(createSelfVote(p2, Neutral))
	g.SetSelfVoteFromPlayer(createSelfVote(p3, LeastVoted))

	roundFinished, _ := g.IsRoundFinished()
	if !roundFinished {
		t.Errorf("expected round to be finished")
	}
	return &g
}

func createTwoVotes(playerName, voteReceiver string) models.PlayerVotedOnQuestion {
	return models.PlayerVotedOnQuestion{
		Player: playerName,
		Votes: []models.Vote{
			{PlayerWhoReceivedTheVote: voteReceiver, QuestionID: "not_used_yet"},
			{PlayerWhoReceivedTheVote: voteReceiver, QuestionID: "not_used_yet"},
		},
	}
}
func createSelfVote(playerName string, selfVote SelfVote) models.RegisterSelfVote {
	return models.RegisterSelfVote{
		Player: playerName,
		Choice: string(selfVote),
		Question: models.Question{
			QuestionID: "not_used_yet",
			Question:   "not_used_yet",
		},
	}
}
