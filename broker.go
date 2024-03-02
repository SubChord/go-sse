package sse

import (
	"github.com/google/uuid"
	"net/http"
	"sync"
	"time"
)

type Broker[idType comparable] struct {
	mtx sync.Mutex

	clientSessions map[idType]map[uuid.UUID]*ClientConnection[idType]
	clientMetadata map[idType]ClientMetadata
	customHeaders  map[string]string

	disconnectCallback func(clientId idType, sessionId uuid.UUID)
}

func NewBroker[idType comparable](customHeaders map[string]string) *Broker[idType] {
	return &Broker[idType]{
		clientSessions: make(map[idType]map[uuid.UUID]*ClientConnection[idType]),
		clientMetadata: map[idType]ClientMetadata{},
		customHeaders:  customHeaders,
	}
}

func (b *Broker[idType]) Connect(clientId idType, w http.ResponseWriter, r *http.Request) (*ClientConnection[idType], error) {
	return b.ConnectWithHeartBeatInterval(clientId, w, r, 15*time.Second)
}

func (b *Broker[idType]) ConnectWithHeartBeatInterval(clientId idType, w http.ResponseWriter, r *http.Request, interval time.Duration) (*ClientConnection[idType], error) {
	client, err := newClientConnection[idType](clientId, w, r)
	if err != nil {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return nil, NewStreamingUnsupportedError(err.Error())
	}

	b.setHeaders(w)

	b.addClient(clientId, client)

	go client.serve(
		interval,
		func() {
			b.removeClient(clientId, client.sessionId) //onClose callback
		},
	)

	return client, nil
}

func (b *Broker[idType]) setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	for k, v := range b.customHeaders {
		w.Header().Set(k, v)
	}
}

func (b *Broker[idType]) IsClientPresent(clientId idType) bool {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	_, ok := b.clientSessions[clientId]
	return ok
}

func (b *Broker[idType]) SetClientMetadata(clientId idType, metadata map[string]interface{}) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	_, ok := b.clientSessions[clientId]
	if !ok {
		return NewUnknownClientError(clientId)
	}

	b.clientMetadata[clientId] = metadata

	return nil
}

func (b *Broker[idType]) GetClientMetadata(clientId idType) (map[string]interface{}, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	_, ok := b.clientSessions[clientId]
	md, ok2 := b.clientMetadata[clientId]
	if !ok || !ok2 {
		return nil, NewUnknownClientError(clientId)
	}

	return md, nil
}

func (b *Broker[idType]) addClient(clientId idType, connection *ClientConnection[idType]) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	_, ok := b.clientSessions[clientId]
	if !ok {
		b.clientSessions[clientId] = make(map[uuid.UUID]*ClientConnection[idType])
	}

	b.clientSessions[clientId][connection.sessionId] = connection
}

func (b *Broker[idType]) removeClient(clientId idType, sessionId uuid.UUID) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	sessions, ok := b.clientSessions[clientId]
	if !ok {
		return
	}

	delete(sessions, sessionId)

	if len(b.clientSessions[clientId]) == 0 {
		delete(b.clientSessions, clientId)
		delete(b.clientMetadata, clientId)
	}

	if b.disconnectCallback != nil {
		go b.disconnectCallback(clientId, sessionId)
	}
}

func (b *Broker[idType]) Broadcast(event Event) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for _, sessions := range b.clientSessions {
		for _, c := range sessions {
			c.Send(event)
		}
	}
}

func (b *Broker[idType]) Send(clientId idType, event Event) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	sessions, ok := b.clientSessions[clientId]
	if !ok {
		return NewUnknownClientError(clientId)
	}
	for _, c := range sessions {
		c.Send(event)
	}
	return nil
}

func (b *Broker[idType]) SetDisconnectCallback(cb func(clientId idType, sessionId uuid.UUID)) {
	b.disconnectCallback = cb
}

func (b *Broker[idType]) Close() error {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	for _, v := range b.clientSessions {
		// Mark all client sessions as done
		for _, session := range v {
			session.doneChan <- struct{}{}
		}
	}

	// Clear client sessions
	b.clientSessions = map[idType]map[uuid.UUID]*ClientConnection[idType]{}

	return nil
}
