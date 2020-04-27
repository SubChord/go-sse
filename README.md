![Go](https://github.com/SubChord/go-sse/workflows/Go/badge.svg?branch=master)

# go-sse
Basic implementation of SSE in golang.
This repository includes a plug and play server-side imlementation and a client-side implementation.
The server-side implementation has been battle-tested while the client-side is usable but in ongoing development.

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
