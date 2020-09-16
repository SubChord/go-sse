package net

import (
	"fmt"
)

type Error struct {
	Code    int
	Message string
	Detail  interface{}
}

var StreamingUnsupportedError = New(1, "streaming unsupported", nil)
var UnknownClientError = New(2, "unknown client", nil)

func New(code int, message string, detail interface{}) Error {
	return Error{Code: code, Message: message, Detail: detail}
}

func (e Error) Error() string {
	return fmt.Sprintf(e.Message)
}
