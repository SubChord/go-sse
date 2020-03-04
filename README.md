![Go](https://github.com/SubChord/go-sse/workflows/Go/badge.svg?branch=master)

# go-sse
Basic implementation of SSE in golang.

## Example usage
```Go
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
	client, err := api.broker.Connect(fmt.Sprintf("%v", time.Now().Unix()), writer, request)
	if err != nil {
		log.Println(err)
		return
	}
	<-client.Done()
	log.Printf("connection with client %v closed", client.Id())
}

```