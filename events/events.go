package events

import (
	"encoding/json"
	"fmt"
	"github.com/akselleirv/introspect/models"
	"github.com/akselleirv/introspect/server"
	"log"
)

func Setup(s server.Server) {
	s.AddEvent("ping", func(data map[string]interface{}) {
		var msg models.Join
		parseToJson(&data, &msg)
		s.SendMsg("test", fmt.Sprintf("hello there %s", msg.Player))
	})
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
