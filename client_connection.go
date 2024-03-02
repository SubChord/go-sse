package sse

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ClientMetadata map[string]interface{}

type ClientConnection[I comparable] struct {
	id        I
	sessionId uuid.UUID

	responseWriter http.ResponseWriter
	request        *http.Request
	flusher        http.Flusher

	msg      chan []byte
	doneChan chan struct{}
}

// newClientConnection is handled by the SSE broker.
func newClientConnection[idType comparable](id idType, w http.ResponseWriter, r *http.Request) (*ClientConnection[idType], error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil, NewStreamingUnsupportedError("ResponseWriter(wrapper) does not support http.Flusher")
	}

	return &ClientConnection[idType]{
		id:             id,
		sessionId:      uuid.New(),
		responseWriter: w,
		request:        r,
		flusher:        flusher,
		msg:            make(chan []byte),
		doneChan:       make(chan struct{}, 1),
	}, nil
}

func (c *ClientConnection[idType]) Id() idType {
	return c.id
}

func (c *ClientConnection[idType]) SessionId() string {
	return c.sessionId.String()
}

func (c *ClientConnection[idType]) Send(event Event) {
	bytes := event.Prepare()
	c.msg <- bytes
}

func (c *ClientConnection[idType]) serve(interval time.Duration, onClose func()) {
	heartBeat := time.NewTicker(interval)

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
				logger.Errorf("unable to write to client %v: %v", c.id, err.Error())
				break writeLoop
			}
			c.flusher.Flush()
		}
	}

	heartBeat.Stop()
	c.doneChan <- struct{}{}
	onClose()
}

func (c *ClientConnection[idType]) Done() <-chan struct{} {
	return c.doneChan
}
