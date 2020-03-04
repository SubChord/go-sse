// +build example3

package main

import (
	"fmt"
	net "github.com/subchord/go-sse"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type API struct {
	broker *net.Broker
}

func main() {
	rand.Seed(time.Now().Unix())

	sseClientBroker := net.NewBroker(map[string]string{
		"Access-Control-Allow-Origin": "*",
	})

	sseClientBroker.SetDisconnectCallback(func(clientId string, sessionId string) {
		log.Printf("session %v of client %v was disconnected.", sessionId, clientId)
	})

	api := &API{broker: sseClientBroker}

	http.HandleFunc("/sse", api.sseHandler)

	go func() {
		ticker := time.Tick(1 * time.Second)
		count := 0
		for {
			select {
			case <-ticker:
				if err := api.broker.Send("always same client", net.StringEvent{
					Id:    fmt.Sprintf("%v", count),
					Event: "message",
					Data:  fmt.Sprintf("%v", count),
				}); err != nil {
					log.Print(err)
				}
				count++
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}

func (api *API) sseHandler(writer http.ResponseWriter, request *http.Request) {
	c, err := api.broker.Connect("always same client", writer, request)
	if err != nil {
		log.Println(err)
		return
	}

	<-c.Done()
}
