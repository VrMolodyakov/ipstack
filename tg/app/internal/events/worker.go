package events

import (
	"app/internal/service/bot"
	"app/pkg/client/mq"
	"app/pkg/logging"
)

type worker struct {
	id         int
	client     mq.Consumer
	producer   mq.Producer
	messages   <-chan mq.Message
	handler    EventHandler
	logger     *logging.Logger
	botService bot.BotService
}

type Worker interface {
	Handle()
}

func NewWorker(
	id int,
	client mq.Consumer,
	handler EventHandler,
	botService bot.BotService,
	producer mq.Producer,
	messages <-chan mq.Message,
	loggeer *logging.Logger) Worker {

	return &worker{
		id:         id,
		client:     client,
		handler:    handler,
		botService: botService,
		messages:   messages,
		producer:   producer, logger: loggeer}
}

func (w *worker) Handle() {
	w.logger.Info("inside")
	for msg := range w.messages {
		w.logger.Info("start parsing message %v", msg)
		eventResponse, err := w.handler.Handle(msg.Body)
		if err != nil {
			w.logger.Errorf("[worker #%d]: failed to processedEvent event due to error %v", w.id, err)
			w.logger.Debugf("[worker #%d]: body: %s", w.id, msg.Body)
			w.reject(msg)
			return
		}
		err = w.botService.SendMessage(eventResponse)
		if err != nil {
			w.logger.Errorf("[worker #%d]: failed to sent message due to error %v", w.id, err)
			w.logger.Debugf("[worker #%d]: body: %s", w.id, msg.Body)
			w.reject(msg)
			return
		}
		w.logger.Info("messages %v was sent", msg)
		w.ack(msg)
	}
	w.logger.Info("exit")
}

func (w *worker) reject(msg mq.Message) {
	if err := w.client.Reject(msg.ID, false); err != nil {
		w.logger.Errorf("[worker #%d]: failed to reject due to error %v", w.id, err)
	}
}

func (w *worker) ack(msg mq.Message) {
	if err := w.client.Ack(msg.ID, false); err != nil {
		w.logger.Errorf("[worker #%d]: failed to ack due to error %v", w.id, err)
	}
}
