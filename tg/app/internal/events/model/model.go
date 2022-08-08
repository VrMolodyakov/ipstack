package model

import "fmt"

type EventResponse struct {
	RequestID string
	Data      map[string]interface{}
	Err       error
}

type ResponseMeta struct {
	RequestID string  `json:"request_id"`
	Error     *string `json:"err,omitempty"`
}

func (e *EventResponse) String() string {
	return fmt.Sprintf("RequestID:%s, Error:%s", e.RequestID, e.Err)
}

func (m *ResponseMeta) String() string {
	return fmt.Sprintf("RequestID:%s, Error:%s", m.RequestID, *m.Error)
}
