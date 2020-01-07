package main

import (
	"bitbucket.org/be-mobile/fleet-notification-api/net"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	sseClientBroker := net.NewBroker(map[string]string{
		"Access-Control-Allow-Origin": "*",
	})

	api := &API{broker: sseClientBroker}

	http.HandleFunc("/sse", api.sseHandler)

	// Broadcast message to all clients every 5 seconds
	go func() {
		count := 0
		tick := time.Tick(5 * time.Second)
		for {
			select {
			case <-tick:
				count++
				api.broker.Broadcast(net.StringEvent{
					Id:    fmt.Sprintf("event-id-%v", count),
					Event: "broadcast_event",
					Data:  strconv.Itoa(count),
				})
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}

type API struct {
	broker *net.Broker
}

func (api *API) sseHandler(writer http.ResponseWriter, request *http.Request) {
	err := api.broker.Connect("c2c6a238-ec23-4700-9fe6-2bcdb393af7b", writer, request)
	if err != nil {
		log.Println(err)
		return
	}
}
