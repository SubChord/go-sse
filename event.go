package net

import (
	"bytes"
	"fmt"
	"strings"
)

type Event interface {
	Prepare() []byte
	GetId() string
	GetEvent() string
	GetData() string
}

type StringEvent struct {
	Id    string
	Event string
	Data  string
}

func (e StringEvent) GetId() string {
	return e.Id
}

func (e StringEvent) GetEvent() string {
	return e.Event
}

func (e StringEvent) GetData() string {
	return e.Data
}

func (e StringEvent) Prepare() []byte {
	var data bytes.Buffer

	if len(e.Id) > 0 {
		data.WriteString(fmt.Sprintf("id: %s\n", strings.Replace(e.Id, "\n", "", -1)))
	}

	data.WriteString(fmt.Sprintf("event: %s\n", strings.Replace(e.Event, "\n", "", -1)))

	// data field should not be empty
	lines := strings.Split(e.Data, "\n")
	for _, line := range lines {
		data.WriteString(fmt.Sprintf("data: %s\n", line))
	}

	data.WriteString("\n")
	return data.Bytes()
}

type HeartbeatEvent struct{}

func (h HeartbeatEvent) GetId() string {
	return ""
}

func (h HeartbeatEvent) GetEvent() string {
	return ""
}

func (h HeartbeatEvent) GetData() string {
	return ""
}

func (h HeartbeatEvent) Prepare() []byte {
	var data bytes.Buffer
	data.WriteString(fmt.Sprint(": heartbeat\n"))
	data.WriteString("\n")
	return data.Bytes()
}
