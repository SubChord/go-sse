package net

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type ClientMetadata map[string]interface{}

type ClientConnection struct {
	id        string
	sessionId string

	responseWriter http.ResponseWriter
	request        *http.Request
	flusher        http.Flusher

	msg      chan []byte
	doneChan chan interface{}
}

// Users should not create instances of client. This should be handled by the SSE broker.
func newClientConnection(id string, w http.ResponseWriter, r *http.Request) (*ClientConnection, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil, NewStreamingUnsupportedError("ResponseWriter(wrapper) does not support http.Flusher")
	}

	return &ClientConnection{
		id:             id,
		sessionId:      uuid.New().String(),
		responseWriter: w,
		request:        r,
		flusher:        flusher,
		msg:            make(chan []byte),
		doneChan:       make(chan interface{}, 1),
	}, nil
}

func (c *ClientConnection) Id() string {
	return c.id
}

func (c *ClientConnection) SessionId() string {
	return c.sessionId
}

func (c *ClientConnection) Send(event Event) {
	bytes := event.Prepare()
	c.msg <- bytes
}

func (c *ClientConnection) serve(interval time.Duration, onClose func()) {
	heartBeat := time.NewTicker(interval)

writeLoop:
	for {
		select {
		case <-c.request.Context().Done():
			close(c.msg)
			break writeLoop
		case <-heartBeat.C:
			go c.Send(HeartbeatEvent{})
		case msg, open := <-c.msg:
			if !open {
				break writeLoop
			}
			_, err := c.responseWriter.Write(msg)
			if err != nil {
				logrus.Errorf("unable to write to client %v: %v", c.id, err.Error())
				break writeLoop
			}
			c.flusher.Flush()
		}
	}

	heartBeat.Stop()
	c.doneChan <- true
	onClose()
}

func (c *ClientConnection) Done() <-chan interface{} {
	return c.doneChan
}
