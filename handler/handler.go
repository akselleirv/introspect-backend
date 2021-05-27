package handler

import (
	"log"
)

type eventHandler = func(msg map[string]interface{})

const Event = "event"

type Handler interface {
	AddEvent(eventName string, fn eventHandler)
	HandleMsg() eventHandler
}

type Handle struct {
	EventHandlers map[string]eventHandler
	l             *log.Logger
}

func NewHandler(l *log.Logger) *Handle {
	return &Handle{EventHandlers: make(map[string]eventHandler), l: l}
}

func (h *Handle) AddEvent(eventName string, fn eventHandler) {
	h.EventHandlers[eventName] = fn
}

func (h *Handle) HandleMsg() func(data map[string]interface{}) {
	return func(msg map[string]interface{}) {
		e, ok := msg[Event].(string)
		if !ok {
			log.Println("sent event is not a string")
			return
		}
		delete(msg, Event)
		handler, ok := h.EventHandlers[e]
		if !ok {
			log.Println("unable to find event in event handlers: ", e)
			return
		}

		h.l.Printf("- %s - %s \n", e, msg)

		handler(msg)
	}
}
