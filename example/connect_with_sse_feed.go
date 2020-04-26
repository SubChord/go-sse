package main

import (
	net "github.com/subchord/go-sse"
	"log"
)

func main() {

	feed, err := net.ConnectWithSSEFeed("http://localhost:8080/sse", nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	subscribe, err := feed.Subscribe("message")
	if err != nil {
		return
	}

	for event := range subscribe.Feed() {
		log.Printf("%v", event)
	}
}
