![Go](https://github.com/SubChord/go-sse/workflows/Go/badge.svg?branch=master)

# go-sse
Basic implementation of SSE in golang.
This repository includes a plug and play server-side imlementation and a client-side implementation.
The server-side implementation has been battle-tested while the client-side is usable but in ongoing development.

Code examples can be found in the `example` folder.
# Install
```bash
go get github.com/SubChord/go-sse@v1.0.0
```
# Server side SSE
1. Create a new broker and pass `optional` headers that should be sent to the client.
```Go
sseClientBroker := net.NewBroker(map[string]string{
	"Access-Control-Allow-Origin": "*",
})
```
2. Set the disconnect callback function if you want to be updated when a client disconnects.
```Go
sseClientBroker.SetDisconnectCallback(func(clientId string, sessionId string) {
	log.Printf("session %v of client %v was disconnected.", sessionId, clientId)
})
```
3. Return an http event stream in an http.Handler. And keep the request open
```Go
func (api *API) sseHandler(writer http.ResponseWriter, request *http.Request) {
	clientConn, err := api.broker.Connect("unique_client_reference", writer, request)
	if err != nil {
		log.Println(err)
		return
	}
	<- clientConn.Done()
}
```
4. After a connection is established you can broadcast events or send client specific events either through the clientConnection instance or through the broker.
```Go
evt := net.StringEvent{
	Id: "self-defined-event-id",
	Event: "type of the event eg. foo_update, bar_delete, ..."
	Data: "data of the event in string format. eg. plain text, json string, ..."
}
api.broker.Broadcast(evt) // all active clients receive this event
api.broker.Send("unique_client_reference", evt) // only the specified client receives this event
&ClientConnection{}.Send(evt) // this instance should be created by the broker only!
```

# Client side SSE
The SSE client makes extensive use of go channels. Once a connection with an SSE feed is established you can subscribe to multiple types of events and process them by looping over the subscription's feed (channel).

1. Connect with SSE feed. And pass `optional` headers.
```Go
headers := make(map[string][]string)
feed, err := net.ConnectWithSSEFeed("http://localhost:8080/sse", headers)
if err != nil {
	log.Fatal(err)
	return
}
```
2. Subscribe to a specific type of event.
```Go
sub, err := feed.Subscribe("message")
if err != nil {
	return
}
```
3. Process the events
```Go
for {
	select {
	case evt := <-sub.Feed():
		log.Print(evt)
	case err := <-sub.ErrFeed():
		log.Fatal(err)
		return
	}
}
```
4. When you are done with all subscriptions and the SSE feed. Don't forget to close the subscriptions and the feed in order to prevent unnecessary network traffic and memory leaks.
```Go
sub.Close()
feed.Close()
```
