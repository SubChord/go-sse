// +build connect_custom_interval

package main

import (
	"fmt"
	"github.com/subchord/go-sse"
	"log"
	"net/http"
	"strconv"
	"time"
)

type API struct {
	broker *net.Broker
}

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
					Event: "message",
					Data:  strconv.Itoa(count),
				})
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}

func (api *API) sseHandler(writer http.ResponseWriter, request *http.Request) {
	// set the heartbeat interval to 1 minute
	client, err := api.broker.ConnectWithHeartBeatInterval(fmt.Sprintf("%v", time.Now().Unix()), writer, request, 1*time.Minute)
	if err != nil {
		log.Println(err)
		return
	}
	<-client.Done()
	log.Printf("connection with client %v closed", client.Id())
}
