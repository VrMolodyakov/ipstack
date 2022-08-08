package events

import (
	"app/internal/events/model"
)

type EventHandler interface {
	Handle(eventBody []byte) (model.EventResponse, error)
}
