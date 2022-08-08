package user

import (
	"context"
	"ipstack/internal/adapters/db/postgresql/client"
	"ipstack/internal/domain/entity"
	psql "ipstack/pkg/client/postgresql"
	"ipstack/pkg/logging"

	sq "github.com/Masterminds/squirrel"
)

const table string = "users"

type UserStorage struct {
	queryBulder sq.StatementBuilderType
	client      client.PostgresClient
	logger      *logging.Logger
}

type UserRepository interface {
	FindAll(context.Context) ([]entity.User, error)
	Insert(context.Context, entity.User) (int, error)
	FindIdByNickname(ctx context.Context, nickname string) (int, error)
}

func NewUserStorage(client client.PostgresClient, logger *logging.Logger) UserRepository {
	return &UserStorage{
		queryBulder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		client:      client,
		logger:      logger,
	}
}

func (u *UserStorage) FindAll(ctx context.Context) ([]entity.User, error) {
	query := u.queryBulder.
		Select("*").
		From(table)
	sql, args, err := query.ToSql()
	logger := psql.QueryLogger(sql, table, *u.logger, args)
	if err != nil {
		err = psql.ErrCreateQuery(err)
		logger.Error(err)
		return nil, err
	}
	logger.Info("process FindAll query")
	rows, err := u.client.Query(ctx, sql, args...)
	if err != nil {
		err = psql.ErrExecuteQuery(err)
		logger.Error(err)
		return nil, err
	}
	defer rows.Close()
	userList := make([]entity.User, 0)
	for rows.Next() {
		user := entity.User{}
		if err = rows.Scan(&user.Id, &user.Nickname); err != nil {
			err = psql.ErrScanRow(err)
			logger.Error(err)
			return nil, err
		}
		userList = append(userList, user)
	}
	return userList, nil
}

func (u *UserStorage) Insert(ctx context.Context, user entity.User) (int, error) {
	sql := `INSERT INTO users(nickname) VALUES($1) RETURNING user_id`
	args := user.Nickname
	u.logger.Info(sql)
	u.logger.Info(args)
	u.logger.Info("process Insert query")
	var userId entity.UserIdDto
	var id int
	err := u.client.QueryRow(ctx, sql, user.Nickname).Scan(&id)
	userId.Id = id
	if err != nil {
		err = psql.ErrExecuteQuery(err)
		u.logger.Error(err)
		return -1, err
	}
	return id, nil
}

func (u *UserStorage) FindIdByNickname(ctx context.Context, nickname string) (int, error) {
	query := u.queryBulder.
		Select("user_id").
		From(table).
		Where(sq.Eq{"nickname": nickname})
	sql, args, err := query.ToSql()
	logger := psql.QueryLogger(sql, table, *u.logger, args)
	if err != nil {
		err = psql.ErrCreateQuery(err)
		logger.Error(err)
		return -1, err
	}
	logger.Info("process Insert query")
	var userId entity.UserIdDto
	var id int
	err = u.client.QueryRow(ctx, sql, args...).Scan(&id)
	userId.Id = id
	return id, err
}
