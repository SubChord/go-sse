package net

import (
	"errors"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type Session struct {
	id string

	responseWriter http.ResponseWriter
	request        *http.Request
	flusher        http.Flusher

	msg      chan []byte
	doneChan chan interface{}
	onClose  func()
	closed   bool
}

// Users should not create instances of client. This should be handled by the SSE broker.
func newSession(w http.ResponseWriter, r *http.Request) (*Session, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	return &Session{
		id:             uuid.New().String(),
		responseWriter: w,
		request:        r,
		flusher:        flusher,
		msg:            make(chan []byte),
		doneChan:       make(chan interface{}),
		onClose:        func() {},
	}, nil
}

func (c *Session) Closed() bool {
	return c.closed
}

func (c *Session) Send(event Event) {
	bytes := event.Prepare()
	c.msg <- bytes
}

func (c *Session) serve(interval time.Duration) error {
	heartBeat := time.NewTicker(interval)
	defer func() {
		heartBeat.Stop()
		c.Close()
		c.onClose()
	}()

writeLoop:
	for {
		select {
		case <-c.request.Context().Done():
			break writeLoop
		case <-heartBeat.C:
			go c.Send(HeartbeatEvent{})
		case msg, open := <-c.msg:
			if !open {
				return errors.New("msg chan closed")
			}
			_, err := c.responseWriter.Write(msg)
			if err != nil {
				return errors.New("write failed")
			}
			c.flusher.Flush()
		}
	}
	return nil
}

func (c *Session) Done() <-chan interface{} {
	return c.doneChan
}

func (c *Session) Close() {
	c.closed = true
	close(c.doneChan)
}
