package internal

import (
	"app/internal/config"
	"app/internal/events"
	"app/internal/events/ipstack"
	"app/internal/service/bot"
	"app/pkg/client/mq"
	"app/pkg/client/mq/rabbitmq"
	"app/pkg/logging"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	tg "gopkg.in/telebot.v3"
)

type app struct {
	cfg          *config.Config
	logger       *logging.Logger
	producer     mq.Producer
	ipstackEvent events.EventHandler
	bot          *tg.Bot
}

type App interface {
	Run()
}

func NewApp(logger *logging.Logger, cfg *config.Config) (App, error) {
	return &app{
		cfg:          cfg,
		logger:       logger,
		ipstackEvent: ipstack.NewIpstackEventHandler(logger),
	}, nil
}

func (a *app) Run() {
	bot, err := a.createBot()
	if err != nil {
		return
	}
	a.bot = bot
	a.startConsume()
	a.bot.Start()
}

func (a *app) startConsume() {
	a.logger.Info("start consuming")
	consumer, err := rabbitmq.NewRabbitMQConsumer(rabbitmq.ConsumerConfig{
		BaseConfig: rabbitmq.BaseConfig{
			Host:     a.cfg.Rabbit.Host,
			Port:     a.cfg.Rabbit.Port,
			Username: a.cfg.Rabbit.Username,
			Password: a.cfg.Rabbit.Password,
		},
		PrefetchCount: a.cfg.Rabbit.Consumer.Buffer,
	})
	if err != nil {
		a.logger.Fatal(err)
	}
	producer, err := rabbitmq.NewRabbitMQProducer(rabbitmq.ProducerConfig{
		BaseConfig: rabbitmq.BaseConfig{
			Host:     a.cfg.Rabbit.Host,
			Port:     a.cfg.Rabbit.Port,
			Username: a.cfg.Rabbit.Username,
			Password: a.cfg.Rabbit.Password,
		},
	})
	if err != nil {
		a.logger.Fatal(err)
	}
	err = consumer.DeclareQueue(a.cfg.Rabbit.Consumer.Ipstack, true, false, false, nil)
	if err != nil {
		a.logger.Fatal(err)
	}
	ipstackMessages, err := consumer.Consume(a.cfg.Rabbit.Consumer.Ipstack)

	if err != nil {
		a.logger.Fatal(err)
	}

	botService := bot.BotService{
		Bot:    a.bot,
		Logger: a.logger,
	}
	a.logger.Info("before")
	a.logger.Info(a.cfg.Event.WorkerCount)

	for i := 0; i < 3; i++ {
		worker := events.NewWorker(i, consumer, a.ipstackEvent, botService, producer, ipstackMessages, a.logger)
		go worker.Handle()
		a.logger.Infof("Ipstack Event Worker #%d started", i)
	}
	a.producer = producer

}

func (a *app) createBot() (bot *tg.Bot, botErr error) {
	tgConfig := tg.Settings{
		Token:   a.cfg.Tg.Token,
		Poller:  &tg.LongPoller{Timeout: 60 * time.Second},
		Verbose: false,
		OnError: a.OnBotError,
	}
	bot, botErr = tg.NewBot(tgConfig)
	if botErr != nil {
		a.logger.Fatal(botErr)
		return
	}

	bot.Handle("/ip", func(c tg.Context) error {
		req := c.Message().Payload
		args := strings.Split(req, " ")
		var request ipstack.IPInfoRequest
		if !isValidIp(args[0]) {
			//return c.Send(fmt.Errorf("incorrect ip address format"))
			request = ipstack.IPInfoRequest{RequestID: fmt.Sprintf("%d", c.Sender().ID), IP: "", Nickname: args[0]}
		} else {
			if len(args) == 1 {
				request = ipstack.IPInfoRequest{RequestID: fmt.Sprintf("%d", c.Sender().ID), IP: args[0], Nickname: ""}
			} else {
				request = ipstack.IPInfoRequest{RequestID: fmt.Sprintf("%d", c.Sender().ID), IP: args[0], Nickname: args[1]}
			}

		}
		marshal, _ := json.Marshal(request)
		err := a.producer.Publish(a.cfg.Rabbit.Producer.Name, marshal)
		if err != nil {
			return c.Send(fmt.Sprintf("failed while publish due to: %s", err.Error()))
		}
		return c.Send(fmt.Sprintf("request accepted"))
	})

	return

}

func (a *app) OnBotError(err error, ctx tg.Context) {
	a.logger.Error(err)
}

func isValidIp(ip string) bool {
	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ip) {
		return true
	}
	return false
}
