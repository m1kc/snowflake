package main

import (
	"fmt"
	"log"
	"time"
)

type hub struct {
	connections map[*connection] bool
	transmit chan message
	register chan *connection
	unregister chan *connection
}

var messageHub = hub {
	connections: make(map[*connection] bool),
	transmit:    make(chan message),
	register:    make(chan *connection),
	unregister:  make(chan *connection),
}

func (h *hub) run() {
	for {
		select {
			case c := <-h.register:
				h.connections[c] = true
				log.Printf("[hub] Registered client")
			case c := <-h.unregister:
				if _, ok := h.connections[c]; ok {
					log.Printf("[hub] Unregistering client")
					delete(h.connections, c)
					close(c.send)
				}
			case m := <-h.transmit:
				for c := range h.connections {
					for ch := range c.channels {
						if c.channels[ch] == m.Channel {
							select {
								case c.send <- m:
								default:
									close(c.send)
									delete(h.connections, c)
							}
						}
					}
				}
		}
	}
}

func (h *hub) generateTestMessages() {
	ticker := time.NewTicker(2 * time.Second)
	i := 1
	for {
		_ = <-ticker.C
		i++
		h.transmit <- message {
			Channel: "broadcast",
			Text: fmt.Sprintf("That's some #broadcast message. Take this number: %d", i),
		}
		h.transmit <- message {
			Channel: "secret",
			Text: fmt.Sprintf("That's some #secret message, woah! Take this number: %d", i*2),
		}
	}
}
