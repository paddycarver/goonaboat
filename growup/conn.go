package main

import (
	"github.com/garyburd/go-websocket/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	readWait       = 60 * time.Second
	pingPeriod     = (readWait * 9) / 10
	maxMessageSize = 512
)

type connection struct {
	ws   *websocket.Conn
	send chan []byte
}

func (c *connection) readPump() {
	defer func() {
		wsHub.unregister <- c
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(readWait))
	for {
		op, r, err := c.ws.NextReader()
		if err != nil {
			break
		}
		switch op {
		case websocket.OpPong:
			c.ws.SetReadDeadline(time.Now().Add(readWait))
		case websocket.OpText:
			message, err := ioutil.ReadAll(r)
			if err != nil {
				break
			}
			log.Println(string(message))
			wsHub.broadcast <- message
			msgHistory.writer <- message
		}
	}
}

func (c *connection) write(opCode int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(opCode, payload)
}

func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.OpClose, []byte{})
				return
			}
			if err := c.write(websocket.OpText, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.OpPing, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	wsHub.register <- c
	go c.writePump()
	c.readPump()
}
