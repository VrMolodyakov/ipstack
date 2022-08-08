package rabbitmq

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type BaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
}

const reconectDelay = 5 * time.Second

var (
	errNotConnected  = errors.New("there is no connectiont to RabbitMQ")
	errAlreadyClosed = errors.New("RabbitMQ connection is already closed")
)

type rabbitMQBase struct {
	lock        sync.Mutex
	isConnected bool
	conn        *amqp.Connection
	ch          *amqp.Channel
	done        chan struct{}
	notifyClose chan *amqp.Error
	reconnects  []chan<- bool
}

func (r *rabbitMQBase) DeclareQueue(name string, durable, autoDelete, exclusive bool, args map[string]interface{}) error {
	if !r.Connected() {
		return errNotConnected
	}
	_, err := r.ch.QueueDeclare(
		name,
		durable,
		autoDelete,
		exclusive,
		false,
		args,
	)
	if err != nil {
		fmt.Errorf("can't declare queue due %v", err)
	}
	return nil
}

func (r *rabbitMQBase) handleReconnect(addr string) {
	for {
		select {
		case <-r.done:
			return
		case err := <-r.notifyClose:
			r.setConnection(false)
			if err != nil {
				return
			}
			log.Println("Trying to reconnect to Rabit MQ")
			for !r.tryToConnect(addr) {
				log.Println("connection establishment failed.Retrying in %v ...", reconectDelay)
				time.Sleep(reconectDelay)
			}
			log.Println("sending signal about siccess reconnoction to Rabbit MQ")
			for _, ch := range r.reconnects {
				ch <- true
			}

		}
	}
}

func (r *rabbitMQBase) notifyIfReconnect(ch chan<- bool) {
	r.reconnects = append(r.reconnects, ch)
}

func (r *rabbitMQBase) tryToConnect(addr string) bool {
	return r.connect(addr) == nil
}

func (r *rabbitMQBase) connect(addr string) error {
	if r.Connected() {
		return nil
	}
	conn, err := amqp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ due %v", err)
	}
	channel, err := conn.Channel()

	if err != nil {
		return fmt.Errorf("failed to open channel due %v", err)
	}
	r.conn = conn
	r.ch = channel
	r.notifyClose = make(chan *amqp.Error)
	r.setConnection(true)
	channel.NotifyClose(r.notifyClose)
	return nil
}

func (r *rabbitMQBase) Connected() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.isConnected
}

func (r *rabbitMQBase) setConnection(isConnected bool) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.isConnected = isConnected
}

func (r *rabbitMQBase) close() error {

	if !r.Connected() {
		return errAlreadyClosed
	}

	if err := r.ch.Close(); err != nil {
		return fmt.Errorf("failed to close the channel due %v", err)
	}

	if err := r.conn.Close(); err != nil {
		return fmt.Errorf("failed to close the connection due %v", err)
	}
	close(r.done)
	r.setConnection(false)
	return nil
}
