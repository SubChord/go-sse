package net

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	neturl "net/url"
	"regexp"
	"strings"
	"sync"
)

type Subscription struct {
	id      string
	parent  *SSEFeed
	feed    chan Event
	errFeed chan error

	eventType string
}

func (s *Subscription) ErrFeed() <-chan error {
	return s.errFeed
}

func (s *Subscription) Feed() <-chan Event {
	return s.feed
}

func (s *Subscription) EventType() string {
	return s.eventType
}

func (s *Subscription) Close() {
	s.parent.closeSubscription(s.id)
}

type SSEFeed struct {
	subscriptions    map[string]*Subscription
	subscriptionsMtx sync.Mutex

	stopChan        chan interface{}
	closed          bool
	unfinishedEvent *StringEvent
}

func ConnectWithSSEFeed(url string, headers map[string][]string) (*SSEFeed, error) {
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    parsedURL,
		Header: headers,
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(resp.Body)

	feed := &SSEFeed{
		subscriptions: make(map[string]*Subscription),
		stopChan:      make(chan interface{}),
	}

	go func(response *http.Response, feed *SSEFeed) {
		defer response.Body.Close()
	loop:
		for {
			select {
			case <-feed.stopChan:
				break loop
			default:
				b, err := reader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					feed.error(err)
					return
				}
				feed.processRaw(b)
			}
		}
	}(resp, feed)

	return feed, nil
}

func (s *SSEFeed) Close() {
	close(s.stopChan)
	for subId, _ := range s.subscriptions {
		s.closeSubscription(subId)
	}
	s.closed = true
}

func (s *SSEFeed) Subscribe(eventType string) (*Subscription, error) {
	if s.closed {
		return nil, fmt.Errorf("sse feed closed")
	}

	sub := &Subscription{
		id:        uuid.New().String(),
		parent:    s,
		eventType: eventType,
		feed:      make(chan Event),
		errFeed:   make(chan error, 1),
	}

	s.subscriptionsMtx.Lock()
	defer s.subscriptionsMtx.Unlock()

	s.subscriptions[sub.id] = sub

	return sub, nil
}

func (s *SSEFeed) closeSubscription(id string) bool {
	s.subscriptionsMtx.Lock()
	defer s.subscriptionsMtx.Unlock()

	if sub, ok := s.subscriptions[id]; ok {
		close(sub.feed)
		return true
	}
	return false
}

func (s *SSEFeed) processRaw(b []byte) {
	if len(b) == 1 && b[0] == '\n' {
		s.subscriptionsMtx.Lock()
		defer s.subscriptionsMtx.Unlock()

		// previous event is complete
		if s.unfinishedEvent == nil {
			return
		}
		evt := StringEvent{
			Id:    s.unfinishedEvent.Id,
			Event: s.unfinishedEvent.Event,
			Data:  s.unfinishedEvent.Data,
		}
		s.unfinishedEvent = nil
		for _, subscription := range s.subscriptions {
            // passing "" gets all eventTypes
			if  subscription.eventType == "" || subscription.eventType == evt.Event {
				subscription.feed <- evt
			}
		}
	}

    split := strings.Split(string(b), "\n")
    var isEvent  = regexp.MustCompile(`^event:\s?`)
    var isId  = regexp.MustCompile(`^id:\s?`)
    var isData  = regexp.MustCompile(`^data:\s?`)

    if s.unfinishedEvent == nil {
		s.unfinishedEvent = &StringEvent{}
	}
    for _, str := range split {
        switch {
        case isEvent.MatchString(str):
            s.unfinishedEvent.Event = isEvent.ReplaceAllString(str,"") 
        case isId.MatchString(str):
            s.unfinishedEvent.Id = isId.ReplaceAllString(str,"")
        case isData.MatchString(str):
            s.unfinishedEvent.Data =  s.unfinishedEvent.Data + isData.ReplaceAllString(str,"\n")
        }
    }
    s.unfinishedEvent.Data = strings.TrimLeft(s.unfinishedEvent.Data,"\n")
}

func (s *SSEFeed) error(err error) {
	s.subscriptionsMtx.Lock()
	defer s.subscriptionsMtx.Unlock()

	for _, subscription := range s.subscriptions {
		subscription.errFeed <- err
	}

	s.Close()
}
