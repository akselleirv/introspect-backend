package events

import (
	"encoding/json"
	"github.com/akselleirv/introspect/handler"
	"github.com/akselleirv/introspect/models"
	"github.com/akselleirv/introspect/room"
	"log"
	"time"
)

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

			err := r.Game().SetPlayerReadyToStartGame(msg.Player)
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
			const event = "get_questions_response"
			var msg models.GenericEvent
			parseToJson(&data, &msg)

			questions, err := r.Game().GetQuestions()
			if err != nil {
				b, _ := json.Marshal(models.ErrorMsg{
					Event:   event,
					Error: err.Error(),
				})
				r.SendMsg(msg.Player, b)
				return
			}
			b, _ := json.Marshal(struct {
				Event     string            `json:"event"`
				Questions []models.Question `json:"questions"`
			}{
				Event:     event,
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
			questionIsDoneMsg := func() ([]byte, error) {
				return json.Marshal(models.QuestionPointsEvent{
					Event:           "question_is_done",
					QuestionPoints:  r.Game().CalculatePointsForCurrentQuestion(),
					CurrentQuestion: r.Game().GetCurrentDoneQuestion(),
				})
			}
			var msg models.RegisterSelfVote
			parseToJson(&data, &msg)
			var b []byte

			r.Game().SetSelfVoteFromPlayer(msg)
			questionDone, allFinished := r.Game().IsRoundFinished()
			log.Println(questionDone, allFinished)
			if allFinished {
				log.Println("game is done")
				b, _ = questionIsDoneMsg()
				log.Println("question is done first res: ", string(b))
				r.Broadcast(b)

				// here we wait for the last question result to be displayed
				// then we send the results for all rounds
				time.Sleep(3 * time.Second)
				cq := r.Game().GetCurrentDoneQuestion()

				b, _ = json.Marshal(models.PlayersResults{
					Event:                        "game_is_finished",
					PlayersResultExceptLastRound: r.Game().CalculatePoints(1, getLastQuestionFromPreviousRound(cq)),
					PlayersResults:               r.Game().CalculatePoints(1, cq),
				})
				log.Println("game is finished: ", string(b))
			} else if questionDone {
				log.Println("all players have self voted for current question")
				b, _ = questionIsDoneMsg()
			} else {
				b, _ = json.Marshal(models.GenericEvent{
					Event:  "player_has_self_voted",
					Player: msg.Player,
				})
			}
			r.Broadcast(b)
		})
		h.AddEvent("next_round", func(data map[string]interface{}) {
			var msg models.GenericEvent
			parseToJson(&data, &msg)
			err := r.Game().SetPlayerReadyForNextRound(msg.Player)
			if err != nil {
				log.Println("unable to find player, removing the player from the game...")
				r.Game().RemovePlayer(msg.Player)
			}

			if r.Game().IsNextRound() {
				b, _ := json.Marshal(models.GenericEvent{
					Event: "all_players_ready_for_next_round",
				})
				r.Broadcast(b)
			}
		})
	}

}

func getLastQuestionFromPreviousRound(currentQuestion int) int {
	return currentQuestion - 4
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
