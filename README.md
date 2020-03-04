![Go](https://github.com/SubChord/go-sse/workflows/Go/badge.svg?branch=master)

# go-sse
Basic implementation of SSE in golang.

## Example usage
```Go
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

	log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
}

func (api *API) sseHandler(writer http.ResponseWriter, request *http.Request) {
	client, err := api.broker.Connect(fmt.Sprintf("%v", rand.Int63()), writer, request)
	if err != nil {
		log.Println(err)
		return
	}

	stop := make(chan interface{}, 1)

	go func() {
		ticker := time.Tick(1 * time.Second)
		count := 0
		for {
			select {
			case <-stop:
				return
			case <-ticker:
				client.Send(net.StringEvent{
					Id:    fmt.Sprintf("%v", count),
					Event: "message",
					Data:  fmt.Sprintf("%v", count),
				})
				count++
			}
		}
	}()

	<-client.Done()
	stop <- true
}
```