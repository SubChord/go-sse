package net

import (
	"github.com/google/uuid"
	"net/http"
	"time"
)

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
		streamingUnsupportedError := StreamingUnsupportedError
		streamingUnsupportedError.Detail = "ResponseWriter(wrapper) does not support http.Flusher"
		return nil, streamingUnsupportedError
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

func (c *ClientConnection) serve(onClose func()) {
	heartBeat := time.NewTicker(15 * time.Second)

writeLoop:
	for {
		select {
		case <-c.request.Context().Done():
			break writeLoop
		case <-heartBeat.C:
			go c.Send(HeartbeatEvent{})
		case msg, open := <-c.msg:
			if !open {
				break writeLoop
			}
			_, err := c.responseWriter.Write(msg)
			if err != nil {
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
