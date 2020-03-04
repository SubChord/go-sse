package net

import (
	"errors"
	"net/http"
	"sync"
)

type Broker struct {
	mtx sync.Mutex

	clients       map[string]*Client
	customHeaders map[string]string

	disconnectCallback func(clientId string)
}

func NewBroker(customHeaders map[string]string) *Broker {
	return &Broker{
		clients:       make(map[string]*Client),
		customHeaders: customHeaders,
	}
}

func (b *Broker) Connect(clientId string, w http.ResponseWriter, r *http.Request) (*Client, error) {
	client, err := NewClient(clientId, w, r)
	if err != nil {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil, errors.New("streaming unsupported")
	}

	b.setHeaders(w)

	b.AddClient(clientId, client)

	go client.serve(
		func() {
			b.RemoveClient(clientId) //onClose callback
		},
	)

	return client, nil
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
	b.mtx.Lock()
	defer b.mtx.Unlock()
	_, ok := b.clients[clientId]
	return ok
}

func (b *Broker) AddClient(clientId string, client *Client) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.clients[clientId] = client
}

func (b *Broker) RemoveClient(clientId string) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	delete(b.clients, clientId)
	if b.disconnectCallback != nil {
		go b.disconnectCallback(clientId)
	}
}

func (b *Broker) Broadcast(event Event) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for _, c := range b.clients {
		c.Send(event)
	}
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
