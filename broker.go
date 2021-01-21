package net

import (
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var ErrUnknownClient = errors.New("client not found")

type Broker struct {
	clientsLock sync.Mutex

	clients       map[string]*Client
	customHeaders map[string]string

	disconnectCallback func(clientId string, sessionId string)
}

type Client struct {
	metadata ClientMetadata
	sessions map[string]*Session
}

type ClientMetadata map[string]interface{}

func NewBroker(customHeaders map[string]string) *Broker {
	return &Broker{
		clients:       make(map[string]*Client),
		customHeaders: customHeaders,
	}
}

func (b *Broker) Connect(clientId string, w http.ResponseWriter, r *http.Request) (*Session, error) {
	return b.ConnectWithHeartBeatInterval(clientId, w, r, 15*time.Second)
}

func (b *Broker) ConnectWithHeartBeatInterval(clientId string, w http.ResponseWriter, r *http.Request, interval time.Duration) (*Session, error) {
	session, err := newSession(w, r)
	if err != nil {
		return nil, err
	}

	session.onClose = func() {
		b.removeClient(clientId, session.id)
	}

	b.setHeaders(w)

	b.addSession(clientId, session)

	go func() {
		err := session.serve(interval)
		if err != nil {
			logrus.Debug(err)
		}
	}()

	return session, nil
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
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	_, ok := b.clients[clientId]
	return ok
}

func (b *Broker) SetClientMetadata(clientId string, metadata map[string]interface{}) error {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	_, ok := b.clients[clientId]
	if !ok {
		return ErrUnknownClient
	}

	b.clients[clientId].metadata = metadata

	return nil
}

func (b *Broker) GetClientMetadata(clientId string) (map[string]interface{}, error) {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	client, ok := b.clients[clientId]
	if !ok {
		return nil, ErrUnknownClient
	}

	return client.metadata, nil
}

func (b *Broker) addSession(clientId string, session *Session) {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	_, ok := b.clients[clientId]
	if !ok {
		b.clients[clientId] = &Client{
			sessions: make(map[string]*Session),
			metadata: ClientMetadata{},
		}
	}

	b.clients[clientId].sessions[session.id] = session
}

func (b *Broker) removeClient(clientId string, sessionId string) {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	clients, ok := b.clients[clientId]
	if !ok {
		return
	}

	delete(clients.sessions, sessionId)

	if len(b.clients[clientId].sessions) == 0 {
		delete(b.clients, clientId)
	}

	if b.disconnectCallback != nil {
		go b.disconnectCallback(clientId, sessionId)
	}
}

func (b *Broker) Broadcast(event Event) {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	for _, client := range b.clients {
		for _, session := range client.sessions {
			session.Send(event)
		}
	}
}

func (b *Broker) Send(clientId string, event Event) error {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	client, ok := b.clients[clientId]
	if !ok {
		return ErrUnknownClient
	}

	for _, c := range client.sessions {
		c.Send(event)
	}

	return nil
}

func (b *Broker) SetDisconnectCallback(cb func(clientId string, sessionId string)) {
	b.disconnectCallback = cb
}

func (b *Broker) Close() error {
	b.clientsLock.Lock()
	defer b.clientsLock.Unlock()

	for _, client := range b.clients {
		// Mark all client sessions as done
		for _, session := range client.sessions {
			session.Close()
		}
	}

	return nil
}
