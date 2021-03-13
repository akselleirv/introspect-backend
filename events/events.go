package events

import (
	"encoding/json"
	"github.com/akselleirv/introspect/handler"
	"github.com/akselleirv/introspect/models"
	"github.com/akselleirv/introspect/room"
	"log"
)

func Setup(h handler.Handler) func(r room.Room) {
	return func(r room.Room) {
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
