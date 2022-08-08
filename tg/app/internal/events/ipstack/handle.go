package ipstack

import (
	"app/internal/events"
	"app/internal/events/model"
	"app/pkg/logging"
	"encoding/json"
	"fmt"
)

type ipstack struct {
	logger *logging.Logger
}

func NewIpstackEventHandler(logger *logging.Logger) events.EventHandler {
	return &ipstack{
		logger: logger,
	}
}

func (i *ipstack) Handle(eventBody []byte) (model.EventResponse, error) {
	event := IPInfoResponse{}
	if err := json.Unmarshal(eventBody, &event); err != nil {
		return model.EventResponse{}, fmt.Errorf("failed  while unmarshaling due %v", err)
	}
	var eventErr error
	if event.Meta.Error != nil {
		eventErr = fmt.Errorf(*event.Meta.Error)
	}
	return model.EventResponse{
		RequestID: event.Meta.RequestID,
		Data:      event.Data.Info,
		Err:       eventErr,
	}, nil
}
