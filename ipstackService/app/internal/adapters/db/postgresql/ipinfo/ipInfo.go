package ipinfo

import (
	"context"
	"ipstack/internal/adapters/db/postgresql/client"
	"ipstack/internal/domain/entity"
	psql "ipstack/pkg/client/postgresql"
	"ipstack/pkg/logging"

	sq "github.com/Masterminds/squirrel"
)

type IPInfoStorage struct {
	queryBulder sq.StatementBuilderType
	client      client.PostgresClient
	logger      *logging.Logger
}

type IPInfoRepository interface {
	FindAll(context.Context) ([]entity.IPInfoDto, error)
	Insert(context.Context, entity.IPInfoDto) (int, error)
	FindIdByIP(ctx context.Context, ip string) (int, error)
}

const table string = "ip_info"

func NewIPInfoStorage(client client.PostgresClient, logger *logging.Logger) IPInfoRepository {
	return &IPInfoStorage{
		queryBulder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		client:      client,
		logger:      logger,
	}
}

func (i *IPInfoStorage) FindAll(ctx context.Context) ([]entity.IPInfoDto, error) {
	query := i.queryBulder.
		Select("*").
		From(table)
	sql, args, err := query.ToSql()
	logger := psql.QueryLogger(sql, table, *i.logger, args)
	if err != nil {
		err = psql.ErrCreateQuery(err)
		logger.Error(err)
		return nil, err
	}
	logger.Info("process FindAll query")
	rows, err := i.client.Query(ctx, sql, args...)
	if err != nil {
		err = psql.ErrExecuteQuery(err)
		logger.Error(err)
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
			logger.Error(err)
			return nil, err
		}
		ipInfoList = append(ipInfoList, info)
	}
	return ipInfoList, nil
}

func (i *IPInfoStorage) Insert(ctx context.Context, info entity.IPInfoDto) (int, error) {
	sql := `INSERT INTO ip_info(ip, continent_name, country_name, region_name, city, zip, latitude, longitude) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING ip_id`
	i.logger.Info("process Insert query")
	var id int
	err := i.client.QueryRow(ctx, sql, info.IP, info.Continent, info.Country, info.Region, info.City, info.Zip, info.Latitude, info.Longitude).Scan(&id)
	if err != nil {
		err = psql.ErrExecuteQuery(err)
		i.logger.Error(err)
		return -1, err
	}
	return id, nil
}

func (i *IPInfoStorage) FindIdByIP(ctx context.Context, ip string) (int, error) {
	query := i.queryBulder.
		Select("ip_id").
		From(table).
		Where(sq.Eq{"ip": ip})
	sql, args, err := query.ToSql()
	logger := psql.QueryLogger(sql, table, *i.logger, args)
	if err != nil {
		err = psql.ErrCreateQuery(err)
		logger.Error(err)
		return -1, err
	}
	logger.Info("process Insert query")
	var id int
	err = i.client.QueryRow(ctx, sql, args...).Scan(&id)
	return id, err
}

// func (i *IPInfoStorage) Insert(ctx context.Context, info entity.IPInfo) (entity.IpIdDto, error) {
// 	query := i.queryBulder.
// 		Insert(table).
// 		Columns("ip", "continent_name", "country_name", "region_name", "city", "zip", "latitude", "longitude").
// 		Values(info.IP, info.Continent, info.Country, info.Region, info.City, info.Zip, info.Latitude, info.Longitude)

// 	sql, args, err := query.ToSql()
// 	logger := psql.QueryLogger(sql, table, *i.logger, args)
// 	if err != nil {
// 		err = psql.ErrCreateQuery(err)
// 		logger.Error(err)
// 		return err
// 	}
// 	logger.Info("process Insert query")
// 	_, err = i.client.Exec(ctx, sql, args...)
// 	if err != nil {
// 		err = psql.ErrExecuteQuery(err)
// 		logger.Error(err)
// 		return err
// 	}
// 	return nil
// }
