package events

import (
	"encoding/json"
	"github.com/akselleirv/introspect/handler"
	"github.com/akselleirv/introspect/models"
	"github.com/akselleirv/introspect/room"
	"log"
	"time"
)

const QuestionsPerRound = 4

func Setup(h handler.Handler) func(r room.Roomer) {
	return func(r room.Roomer) {
		h.AddEvent("ping", func(data map[string]interface{}) {
			var msg models.Ping
			parseToJson(&data, &msg)
			ping := models.Ping{
				Event:  "ping",
				Player: msg.Player,
			}
			b, _ := json.Marshal(ping)
			r.SendMsg(msg.Player, b)
		})
		h.AddEvent("ping_broadcast", func(data map[string]interface{}) {
			var msg models.Ping
			parseToJson(&data, &msg)
			ping := models.Ping{
				Event:  "ping_broadcast",
				Player: msg.Player,
			}
			b, _ := json.Marshal(ping)
			r.Broadcast(b)
		})
		h.AddEvent("lobby_chat", func(data map[string]interface{}) {
			var msg models.LobbyChat
			parseToJson(&data, &msg)
			res := models.LobbyChat{
				Event:   "lobby_chat",
				Player:  msg.Player,
				Message: msg.Message,
			}
			b, _ := json.Marshal(res)
			r.Broadcast(b)
		})
		h.AddEvent("lobby_player_ready", func(data map[string]interface{}) {
			var msg models.GenericEvent
			parseToJson(&data, &msg)

			err := r.Game().SetPlayerReady(msg.Player)
			if err != nil {
				// TODO: Handle error when setting player ready
				return
			}

			playersUpdate, isAllReady := r.Game().GetRoomStatus()
			b, _ := json.Marshal(models.LobbyRoomUpdate{
				Event:      "lobby_room_update",
				Players:    playersUpdate,
				IsAllReady: isAllReady,
			})
			r.Broadcast(b)
		})
		h.AddEvent("get_questions_request", func(data map[string]interface{}) {
			var msg models.GenericEvent
			parseToJson(&data, &msg)

			questions := r.Game().GetQuestions()
			b, _ := json.Marshal(struct {
				Event     string            `json:"event"`
				Questions []models.Question `json:"questions"`
			}{
				Event:     "get_questions_response",
				Questions: questions,
			})
			r.SendMsg(msg.Player, b)
		})
		h.AddEvent("register_question_vote", func(data map[string]interface{}) {
			var msg models.PlayerVotedOnQuestion
			parseToJson(&data, &msg)
			var b []byte
			r.Game().SetVotesFromPlayer(msg)
			isSelfVoting := r.Game().IsSelfVoting()
			if isSelfVoting {
				b, _ = json.Marshal(models.GenericEvent{
					Event:  "is_self_vote",
					Player: "",
				})
			} else {
				b, _ = json.Marshal(models.GenericEvent{
					Event:  "player_has_question_voted",
					Player: msg.Player,
				})
			}
			r.Broadcast(b)
		})
		h.AddEvent("register_self_vote", func(data map[string]interface{}) {
			var msg models.RegisterSelfVote
			parseToJson(&data, &msg)
			var b []byte

			r.Game().SetSelfVoteFromPlayer(msg)
			questionDone, allFinished := r.Game().IsRoundFinished()
			log.Println(questionDone, allFinished)
			if allFinished {
				log.Println("game is done")
				b, _ = json.Marshal(models.QuestionPointsEvent{
					Event:           "question_is_done",
					QuestionPoints:  r.Game().CalculatePointsForCurrentQuestion(),
					CurrentQuestion: r.Game().GetCurrentQuestion(),
				})
				r.Broadcast(b)

				// here we wait for the last question result to be displayed
				// then we send the results for all rounds
				time.Sleep(5 * time.Second)
				currQ := r.Game().GetCurrentQuestion()

				b, _ = json.Marshal(models.PlayersResultsTotal{
					Event:                  "game_is_finished",
					PlayersResultLastRound: r.Game().CalculatePoints(currQ-QuestionsPerRound+1, currQ),
					PlayersResultsTotal:    r.Game().CalculatePoints(1, currQ),
				})
			} else if questionDone {
				log.Println("all players have self voted for current question")
				b, _ = json.Marshal(models.QuestionPointsEvent{
					Event:           "question_is_done",
					QuestionPoints:  r.Game().CalculatePointsForCurrentQuestion(),
					CurrentQuestion: r.Game().GetCurrentQuestion(),
				})
			} else {
				b, _ = json.Marshal(models.GenericEvent{
					Event:  "player_has_self_voted",
					Player: msg.Player,
				})
			}
			r.Broadcast(b)
		})
	}

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
