package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"strings"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader {
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type connection struct {
	ws *websocket.Conn
	send chan message
	id string
	channels []string
}

type message struct {
	Channel string
	Text string
}

func (c *connection) readPump() {
	defer func() {
		messageHub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, body, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		msg := string(body)
		msgFields := strings.SplitN(msg, ",", 3)
		switch msgFields[0] {
			case "id":
				log.Printf("[conn] Client told ID: %s", msgFields[1])
				c.id = msgFields[1]
			case "sub":
				c.channels = strings.Fields(msgFields[1])
				log.Printf("[conn] Client joined %d channel(s): %v", len(c.channels), c.channels)
			case "send":
				log.Printf("[conn] Transmission to #%s: %s", msgFields[1], msgFields[2])
				messageHub.transmit <- message {
					Channel: msgFields[1],
					Text: msgFields[2],
				}
			default:
				log.Printf("[conn] Can't parse incoming message: %s", msg)
		}
		/*
		messageHub.transmit <- message {
			channel: "broadcast",
			text: body,
		}
		*/
	}
}

func (c *connection) write(mt int, payload string) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, []byte(payload))
}

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
			case msg, ok := <-c.send:
				if !ok {
					c.write(websocket.CloseMessage, "")
					return
				}
				tmp, _ := json.Marshal(&msg)
				if err := c.write(websocket.TextMessage, string(tmp)); err != nil {
					return
				}
			case <-ticker.C:
				if err := c.write(websocket.PingMessage, ""); err != nil {
					return
				}
			}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection {
		send: make(chan message, 256),
		ws: ws,
		id: "unknown",
		channels: []string{"broadcast"},
	}
	messageHub.register <- c
	go c.writePump()
	c.readPump()
}
