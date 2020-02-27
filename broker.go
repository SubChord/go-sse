package net

import (
	"errors"
	"net/http"
	"sync"
)

type Broker struct {
	mtx sync.Mutex

	clients       map[string]*client
	customHeaders map[string]string

	disconnectCallback func(clientId string)
}

func NewBroker(customHeaders map[string]string) *Broker {
	return &Broker{
		clients:       make(map[string]*client),
		customHeaders: customHeaders,
	}
}

func (b *Broker) Connect(clientId string, w http.ResponseWriter, r *http.Request) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return errors.New("streaming unsupported")
	}

	client := &client{
		id:   clientId,
		conn: w,
		msg:  make(chan []byte),
	}

	b.setHeaders(w)

	b.AddClient(clientId, client)
	client.Serve(func() {
		flusher.Flush() // write callback
	})
	b.RemoveClient(clientId)

	return nil
}

func (b *Broker) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	for k, v := range b.customHeaders {
		w.Header().Set(k, v)
	}
}

func (b *Broker) IsClientPresent(clientId string) bool {
	_, ok := b.clients[clientId]
	return ok
}

func (b *Broker) AddClient(clientId string, client *client) {
	b.mtx.Lock()
	b.clients[clientId] = client
	b.mtx.Unlock()
}

func (b *Broker) RemoveClient(clientId string) {
	b.mtx.Lock()
	delete(b.clients, clientId)
	if b.disconnectCallback != nil {
		go b.disconnectCallback(clientId)
	}
	b.mtx.Unlock()
}

func (b *Broker) Broadcast(event Event) {
	b.mtx.Lock()
	for _, c := range b.clients {
		c.Send(event)
	}
	b.mtx.Unlock()
}

func (b *Broker) Send(clientId string, event Event) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	c, ok := b.clients[clientId]
	if !ok {
		return errors.New("unknown client")
	}
	c.Send(event)
	return nil
}

func (b *Broker) SetDisconnectCallback(cb func(clientId string)) {
	b.disconnectCallback = cb
}
