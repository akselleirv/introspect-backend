package client

import (
	"github.com/gorilla/websocket"
	"log"
)

type Clienter interface {
	Send(msg []byte)
}

type Client struct {
	name  string
	msgCh chan<- []byte
}

func NewClient(name string, c *websocket.Conn, msgHandler func(msg map[string]interface{}), removeClientFromRoom func()) *Client {
	ch := make(chan []byte)
	go readMessages(c, msgHandler, removeClientFromRoom)
	go writeMessages(c, ch, removeClientFromRoom)

	return &Client{
		name:  name,
		msgCh: ch,
	}
}

func (c *Client) Send(msg []byte) {
	c.msgCh <- msg
}

// readMessages reads messages from conn and sends the msg to the handler
func readMessages(c *websocket.Conn, msgHandler func(msg map[string]interface{}), removeClientFromRoom func()) {
	msg := make(map[string]interface{})
	for {
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			removeClientFromRoom()
			break
			//TODO: Set a timer in order to move inactive clients
		}
		log.Printf("recv: %s", msg)

		msgHandler(msg)
	}
}

// writeMessages writes to the Client by reading from the msgCh
func writeMessages(c *websocket.Conn, msgToSend <-chan []byte, removeClientFromRoom func()) {
	var err error
	for m := range msgToSend {
		err = c.WriteMessage(1, m)
		if err != nil {
			log.Println("unable to write to client: ", err)
			removeClientFromRoom()
			break
		}
	}
}
