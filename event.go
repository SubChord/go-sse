package net

import (
	"bytes"
	"fmt"
	"strings"
)

type Event interface {
	Prepare() []byte
}

type StringEvent struct {
	Id    string
	Event string
	Data  string
}

func (m StringEvent) Prepare() []byte {
	var data bytes.Buffer
	if len(m.Id) > 0 {
		data.WriteString(fmt.Sprintf("id: %s\n", strings.Replace(m.Id, "\n", "", -1)))
	}
	if len(m.Event) > 0 {
		data.WriteString(fmt.Sprintf("event: %s\n", strings.Replace(m.Event, "\n", "", -1)))
	}
	if len(m.Data) > 0 {
		lines := strings.Split(m.Data, "\n")
		for _, line := range lines {
			data.WriteString(fmt.Sprintf("data: %s\n", line))
		}
	}
	data.WriteString("\n")
	return data.Bytes()
}

type HeartbeatEvent struct {}

func (m HeartbeatEvent) Prepare() []byte {
	var data bytes.Buffer
	data.WriteString(fmt.Sprint(": heartbeat\n"))
	data.WriteString("\n")
	return data.Bytes()
}
