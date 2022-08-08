package bot

import (
	"app/internal/events/model"
	"app/pkg/logging"
	"bytes"
	"fmt"
	"strconv"

	tg "gopkg.in/telebot.v3"
)

type BotService struct {
	Bot    *tg.Bot
	Logger *logging.Logger
}

func (bs *BotService) SendMessage(data model.EventResponse) error {
	reqId, _ := strconv.ParseInt(data.RequestID, 10, 64)
	id, err := bs.Bot.ChatByID(reqId)
	if err != nil {
		bs.Logger.Tracef("bot send responseMessage. ProcessedEvent: %s", &data)
		return fmt.Errorf("failed to get chat by id due to %v", err)
	}
	message := data.Data
	if data.Err != nil {
		message = map[string]interface{}{"error": data.Err}
	}
	response := mapResponseToString(message)
	_, err = bs.Bot.Send(id, response)
	if err != nil {
		bs.Logger.Tracef("ChatID: %d, Data: %s", id.ID, data)
		return fmt.Errorf("failed while sending due to  %v", err)
	}
	return nil
}

func mapResponseToString(r map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range r {
		fmt.Fprintf(b, "%s=\"%v\"\n", key, value)
	}
	return b.String()
}
