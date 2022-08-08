package internal

import (
	"context"
	"ipstack/internal/adapters/db/postgresql/user"
	"ipstack/internal/config"
	"ipstack/internal/controller/http/ipstack"
	v1 "ipstack/internal/controller/http/v1"
	"ipstack/internal/domain/service"
	"ipstack/internal/events"
	"ipstack/pkg/client/mq/rabbitmq"
	"ipstack/pkg/client/postgresql"
	"ipstack/pkg/logging"

	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type app struct {
	cfg           *config.Config
	logger        *logging.Logger
	http          *gin.Engine
	ipService     ipstack.HttpService
	userIpService service.UserIPInfoSerivce
}

type App interface {
	Run()
}

func NewApp(logger *logging.Logger, cfg *config.Config, http *gin.Engine, ipService ipstack.HttpService, userIpService service.UserIPInfoSerivce) (App, error) {
	return &app{
		cfg:           cfg,
		logger:        logger,
		http:          http,
		ipService:     ipService,
		userIpService: userIpService,
	}, nil
}

func (a *app) Run() {
	a.startConsume()
	//a.startHttp()
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

	err = consumer.DeclareQueue(a.cfg.Rabbit.Producer.Name, true, false, false, nil)
	if err != nil {
		a.logger.Fatal(err)
	}
	messages, err := consumer.Consume(a.cfg.Rabbit.Producer.Name)
	if err != nil {
		a.logger.Fatal(err)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		worker := events.NewWorker(i, consumer, a.cfg.Rabbit.Consumer.Ipstack, producer, messages, a.logger, a.ipService, a.userIpService, &wg)
		wg.Add(1)
		go worker.Process()
		a.logger.Infof("Event Worker #%d started", i)
	}

	wg.Wait()
}

func (a *app) startHttp() {
	pgConfig := postgresql.NewPgConfig(
		a.cfg.Postgresqldb.Username,
		a.cfg.Postgresqldb.Password,
		a.cfg.Postgresqldb.Host,
		a.cfg.Postgresqldb.Port,
		a.cfg.Postgresqldb.Dbname)

	pgClient, err := postgresql.NewClient(context.Background(), 5, time.Second*5, pgConfig)
	if err != nil {
		a.logger.Fatal(err)
	}
	userRepo := user.NewUserStorage(pgClient, a.logger)
	userService := service.NewUserService(userRepo)
	userHandler := v1.NewUserHandler(userService)

	a.http.POST("/user", userHandler.CreateUser)
	a.http.GET("/users", userHandler.GetAllUsers)
	a.http.Run()
}
