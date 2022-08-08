package main

import (
	"context"
	"ipstack/internal"
	"ipstack/internal/adapters/db/postgresql/ipinfo"
	"ipstack/internal/adapters/db/postgresql/user"
	useripinfo "ipstack/internal/adapters/db/postgresql/userIPInfo"
	"ipstack/internal/config"
	"ipstack/internal/controller/http/ipstack"
	"ipstack/internal/domain/service"
	"ipstack/pkg/client/postgresql"
	"ipstack/pkg/logging"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	//dbUrl := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", "postgres", "Loken", "localhost", "devdb")
	log.Print("config initializing")
	cfg := config.GetConfig()
	log.Print("logger initializing")
	logger := logging.GetLogger("info")
	logger.Println("Creating Application")
	ipService := ipstack.NewHttpService(cfg.Ipstack.Key, logger)

	pgConfig := postgresql.NewPgConfig(
		cfg.Postgresqldb.Username,
		cfg.Postgresqldb.Password,
		cfg.Postgresqldb.Host,
		cfg.Postgresqldb.Port,
		cfg.Postgresqldb.Dbname)

	pgClient, err := postgresql.NewClient(context.Background(), 5, time.Second*5, pgConfig)
	if err != nil {
		logger.Fatal(err)
	}
	userRepo := user.NewUserStorage(pgClient, logger)
	ipRepo := ipinfo.NewIPInfoStorage(pgClient, logger)
	userIpRepo := useripinfo.NewUserStorage(pgClient, logger)
	userService := service.NewUserService(userRepo)
	ipInfoService := service.NewIPInfoService(ipRepo)
	userIpInfoService := service.NewUserIPInfoService(userIpRepo, userService, ipInfoService)

	// users, err := userRepo.FindAll(context.Background())
	// if err != nil {
	// 	logger.Fatal(err)
	// }
	// ips, err := ipRepo.FindAll(context.Background())
	// if err != nil {
	// 	logger.Fatal(err)
	// }
	// logger.Info(users)
	// logger.Info(ips)

	app, err := internal.NewApp(logger, cfg, gin.Default(), *ipService, userIpInfoService)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Println("Running Application")
	app.Run()
}
