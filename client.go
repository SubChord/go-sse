package net

import (
	"net/http"
)

type client struct {
	id   string
	conn http.ResponseWriter
	msg  chan []byte
}

func (c *client) Send(event Event) {
	c.msg <- event.Prepare()
}

func (c *client) Serve(writeCallback func()) {
	for {
		select {
		case msg, open := <-c.msg:
			if !open {
				return
			}
			_, err := c.conn.Write(msg)
			if err != nil {
				return
			}
			writeCallback()
		}
	}
}
