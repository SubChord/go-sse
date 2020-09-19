// +build connect_with_sse

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

	sub, err := feed.Subscribe("message")
	if err != nil {
		return
	}

	for {
		select {
		case evt := <-sub.Feed():
			log.Print(evt)
		case err := <-sub.ErrFeed():
			log.Fatal(err)
			return
		}
	}

	sub.Close()
	feed.Close()
}
