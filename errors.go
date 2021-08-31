package net

import "fmt"

type SSEError struct {
	Message string
}

type StreamingUnsupportedError struct {
	SSEError
}

func NewStreamingUnsupportedError(msg string) *StreamingUnsupportedError {
	return &StreamingUnsupportedError{SSEError: SSEError{Message: msg}}
}

func (s StreamingUnsupportedError) Error() string {
	return s.Message
}

type UnknownClientError struct {
	SSEError
}

func NewUnknownClientError(clientId string) *UnknownClientError {
	return &UnknownClientError{SSEError{Message: fmt.Sprintf("clientId: %v", clientId)}}
}

func (u UnknownClientError) Error() string {
	return u.Message
}
