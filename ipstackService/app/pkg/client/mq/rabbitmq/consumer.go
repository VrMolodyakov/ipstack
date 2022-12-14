package rabbitmq

import (
	"fmt"
	"ipstack/pkg/client/mq"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type ConsumerConfig struct {
	BaseConfig
	//max message count at the time
	PrefetchCount int
}

type rabbitMQConsumer struct {
	*rabbitMQBase
	reconnectCh   chan bool
	prefetchCount int
}

const consumeDelay = 1 * time.Second

func NewRabbitMQConsumer(cfg ConsumerConfig) (mq.Consumer, error) {
	consumer := &rabbitMQConsumer{
		prefetchCount: cfg.PrefetchCount,
		reconnectCh:   make(chan bool),
		rabbitMQBase: &rabbitMQBase{
			done: make(chan struct{}),
		},
	}
	addr := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	err := consumer.connect(addr)
	if err != nil {
		return nil, err
	}
	consumer.notifyIfReconnect(consumer.reconnectCh)
	go consumer.handleReconnect(addr)
	return consumer, nil
}

func (r *rabbitMQConsumer) Consume(target string) (<-chan mq.Message, error) {
	if !r.Connected() {
		return nil, errNotConnected
	}
	messages, err := r.consume(target)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages due %v", err)
	}
	messageChan := make(chan mq.Message)
	go func() {
		for {
			select {
			case message, ok := <-messages:
				if !ok {
					time.Sleep(consumeDelay)
					continue
				}
				messageChan <- mq.Message{
					ID:   message.DeliveryTag,
					Body: message.Body,
				}
			case <-r.reconnectCh:
				log.Println("start to reconsume messages")
				for {
					messages, err = r.consume(target)
					if err == nil {
						break
					}

					log.Printf("failed to reconsume messages due %v", err)
				}
			case <-r.done:
				close(messageChan)
				return
			}
		}
	}()
	return messageChan, nil
}

func (r *rabbitMQConsumer) consume(target string) (<-chan amqp.Delivery, error) {
	err := r.ch.Qos(r.prefetchCount, 0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to set Qos due %v", err)
	}
	messages, err := r.ch.Consume(
		target,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *rabbitMQConsumer) Ack(id uint64, multiple bool) error {
	if !r.Connected() {
		return errNotConnected
	}
	err := r.ch.Ack(id, multiple)
	if err != nil {
		return fmt.Errorf("failed to ack message with id: %d due %v", id, err)
	}
	return nil
}

func (r *rabbitMQConsumer) Nack(id uint64, multiple bool, requeue bool) error {
	if !r.Connected() {
		return errNotConnected
	}
	err := r.ch.Nack(id, multiple, requeue)

	if err != nil {
		return fmt.Errorf("failed while nack message with id: %d due %v", id, err)
	}
	return nil
}

func (r *rabbitMQConsumer) Reject(id uint64, requeue bool) error {
	if !r.Connected() {
		return errNotConnected
	}
	err := r.ch.Reject(id, requeue)

	if err != nil {
		return fmt.Errorf("failed while rejecting message with id: %d due %v", id, err)
	}
	return nil
}

func (r *rabbitMQConsumer) Close() error {
	if err := r.close(); err != nil {
		return err
	}
	return nil
}
