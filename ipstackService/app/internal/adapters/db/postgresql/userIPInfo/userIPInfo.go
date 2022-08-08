package useripinfo

import (
	"context"
	"ipstack/internal/adapters/db/postgresql/client"
	"ipstack/internal/domain/entity"
	psql "ipstack/pkg/client/postgresql"
	"ipstack/pkg/logging"

	sq "github.com/Masterminds/squirrel"
)

const table string = "user_ip_info"

type UserIPInfoStorage struct {
	queryBulder sq.StatementBuilderType
	client      client.PostgresClient
	logger      *logging.Logger
}

type UserIPInfoRepository interface {
	FindAll(context.Context) ([]entity.UserIPInfo, error)
	Insert(context.Context, entity.UserIPInfo) error
	FindUserIpAddresess(ctx context.Context, userId int) ([]entity.IPInfoDto, error)
}

func NewUserStorage(client client.PostgresClient, logger *logging.Logger) UserIPInfoRepository {
	return &UserIPInfoStorage{
		queryBulder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		client:      client,
		logger:      logger,
	}
}

func (u *UserIPInfoStorage) FindAll(ctx context.Context) ([]entity.UserIPInfo, error) {
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
	userList := make([]entity.UserIPInfo, 0)
	for rows.Next() {
		user := entity.UserIPInfo{}
		if err = rows.Scan(&user.Id, &user.UserId, &user.IPId); err != nil {
			err = psql.ErrScanRow(err)
			logger.Error(err)
			return nil, err
		}
		userList = append(userList, user)
	}
	return userList, nil
}

/*
	insert into user_ip_info("user_id", "ip_id")
	select $1,$2
	where not exists (select user_id", "ip_id from user_ip_info)



*/
func (u *UserIPInfoStorage) Insert(ctx context.Context, userIPInfo entity.UserIPInfo) error {

	sql := `INSERT INTO user_ip_info(user_id, ip_id)
		 	SELECT $1,$2
			WHERE NOT EXISTS (SELECT user_id, ip_id FROM user_ip_info WHERE user_id=$1 AND ip_id=$2)`

	u.logger.Info("process Insert query")
	_, err := u.client.Exec(ctx, sql, userIPInfo.UserId, userIPInfo.IPId)
	if err != nil {
		err = psql.ErrExecuteQuery(err)
		u.logger.Error(err)
		return err
	}
	return nil
}

func (u *UserIPInfoStorage) FindUserIpAddresess(ctx context.Context, userId int) ([]entity.IPInfoDto, error) {
	sql := `SELECT ip, continent_name, country_name, region_name, city, zip, latitude, longitude
			FROM user_ip_info
			INNER JOIN ip_info ON  user_ip_info.ip_id=ip_info.ip_id
			WHERE user_id = $1`
	u.logger.Info("process Insert query")
	rows, err := u.client.Query(ctx, sql, userId)
	if err != nil {
		err = psql.ErrExecuteQuery(err)
		u.logger.Error(err)
		return nil, err
	}
	defer rows.Close()
	ipInfoList := make([]entity.IPInfoDto, 0)
	for rows.Next() {
		info := entity.IPInfoDto{}
		if err = rows.Scan(
			&info.IP,
			&info.Continent,
			&info.Country,
			&info.Region,
			&info.City,
			&info.Zip,
			&info.Latitude,
			&info.Longitude); err != nil {
			err = psql.ErrScanRow(err)
			u.logger.Error(err)
			return nil, err
		}
		ipInfoList = append(ipInfoList, info)
	}
	return ipInfoList, nil
}
