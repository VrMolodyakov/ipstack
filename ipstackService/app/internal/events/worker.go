package events

import (
	"context"
	"encoding/json"
	"fmt"
	"ipstack/internal/controller/http/ipstack"
	"ipstack/internal/domain/entity"
	"ipstack/internal/domain/service"
	"ipstack/internal/events/model/request"
	"ipstack/internal/events/model/response"
	"ipstack/pkg/client/mq"
	"ipstack/pkg/logging"
	"sync"
)

type Worker struct {
	id            int
	client        mq.Consumer
	producer      mq.Producer
	responseQueue string
	messages      <-chan mq.Message
	logger        *logging.Logger
	ipstack       ipstack.HttpService
	userIpService service.UserIPInfoSerivce
	wg            *sync.WaitGroup
}

func NewWorker(id int, client mq.Consumer, responseQueue string, producer mq.Producer, messages <-chan mq.Message,
	logger *logging.Logger, ipstack ipstack.HttpService, userIpService service.UserIPInfoSerivce, wg *sync.WaitGroup) Worker {

	return Worker{id: id, client: client, responseQueue: responseQueue, messages: messages, producer: producer,
		logger: logger, ipstack: ipstack, userIpService: userIpService, wg: wg}
}

func (w *Worker) Process() {
	defer w.wg.Done()
	for msg := range w.messages {
		w.logger.Info(msg)
		w.logger.Info("sending")
		event := request.IpSearchRequest{}
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			w.logger.Errorf("[Worker #%d]: failed to unmarchal event due to error %v", w.id, err)
			w.logger.Debugf("[Worker #%d]: body: %s", w.id, msg.Body)

			w.reject(msg)
			continue
		}
		var infoDto entity.IPInfoDto
		var err error
		if event.IP != "" {
			infoDto, err = w.ipstack.GetIPInfo(context.Background(), event.IP)
			if err != nil {
				e := fmt.Sprintf("%v", err)
				r := response.Meta{RequestID: event.RequestID, Error: &e}
				w.sendResponse(response.IPResponse{Meta: r})
			}
			w.convertAndSend(infoDto, event)
		}
		if event.IP == "" && event.Nickname != "" {
			ipAddresess, err := w.userIpService.GetAllUserIpInfo(context.Background(), event.Nickname)
			if err == nil {
				for _, ipInfo := range ipAddresess {
					w.convertAndSend(ipInfo, event)
				}
			}
		}

		if len(event.Nickname) != 0 && len(event.IP) != 0 {
			w.userIpService.Create(context.Background(), event.Nickname, infoDto)
		}
		w.ack(msg)
	}
	w.logger.Info("end")
}

func (w *Worker) convertAndSend(ipInfo entity.IPInfoDto, event request.IpSearchRequest) {
	data, err := json.Marshal(ipInfo) // Convert to a json string
	if err != nil {
		w.logger.Error(err)
	}
	respMap := make(map[string]interface{})
	err = json.Unmarshal(data, &respMap)
	if err != nil {
		w.logger.Error(err)
	}
	response := response.IPResponse{Meta: response.Meta{RequestID: event.RequestID}, Data: response.Data{respMap}}
	w.sendResponse(response)
}

func (w *Worker) sendResponse(d interface{}) {
	b, err := json.Marshal(d)
	if err != nil {
		w.logger.Errorf("[Worker #%d]: failed to response due to error %v", w.id, err)
		return
	}
	err = w.producer.Publish(w.responseQueue, b)
	if err != nil {
		w.logger.Errorf("[Worker #%d]: failed to response due to error %v", w.id, err)
	}
}

func (w *Worker) reject(msg mq.Message) {
	if err := w.client.Reject(msg.ID, false); err != nil {
		w.logger.Errorf("[Worker #%d]: failed to reject due to error %v", w.id, err)
	}
}

func (w *Worker) ack(msg mq.Message) {
	if err := w.client.Ack(msg.ID, false); err != nil {
		w.logger.Errorf("[Worker #%d]: failed to ack due to error %v", w.id, err)
	}
}
