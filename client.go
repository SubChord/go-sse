package net

import (
	"net/http"
)

type Client struct {
	id             string
	responseWriter http.ResponseWriter
	request        *http.Request
	msg            chan []byte
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) Send(event Event) {
	c.msg <- event.Prepare()
}

func (c *Client) serve(onWrite func(), onClose func()) {
writeLoop:
	for {
		select {
		case <-c.request.Context().Done():
			break writeLoop
		case msg, open := <-c.msg:
			if !open {
				break writeLoop
			}
			_, err := c.responseWriter.Write(msg)
			if err != nil {
				return
			}
			onWrite()
		}
	}
	onClose()
}
