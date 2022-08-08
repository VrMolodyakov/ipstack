package postgresql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Client interface {
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginFunc(ctx context.Context, f func(pgx.Tx) error) error
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	BeginTxFunc(ctx context.Context, txOption pgx.TxOptions, f func(pgx.Tx) error) error
}

type pgConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

func NewPgConfig(username string, password string, host string, port string, database string) *pgConfig {
	return &pgConfig{
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		Database: database,
	}
}

func NewClient(ctx context.Context, maxAttempts int, delay time.Duration, cfg *pgConfig) (pool *pgxpool.Pool, err error) {
	connectUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	err = DoWithAttempts(func() error {
		//ctx, cansel := context.WithTimeout(ctx, 10*time.Second)
		//defer cansel()
		config, err := pgxpool.ParseConfig(connectUrl)
		if err != nil {
			log.Fatalf("Failed while parsing config: %v\n", err)
		}
		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err != nil {
			log.Println("Connection failed... Going to do the next attempt")

			return err
		}
		return nil
	}, maxAttempts, delay)
	if err != nil {
		log.Fatal("cannot connect to Postgresql")
	}
	log.Println("Connection has been established")
	return pool, nil

}

func DoWithAttempts(f func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = f(); err != nil {
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil

	}
	return
}

// connectUrl := fmt.Sprintf("postgres://%s:%s@%s:5432/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Dbname)
// 	err = DoWithAttempts(func() error {
// 		ctx, cansel := context.WithTimeout(ctx, 3*time.Second)
// 		defer cansel()
// 		pool, err = pgxpool.Connect(ctx, connectUrl)
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}, maxAttempts, 3*time.Second)
// 	if err != nil {
// 		log.Fatal("cannot connect to Postgresql")
// 	}
// 	return pool, nil
