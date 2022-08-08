package service

import (
	"context"
	"ipstack/internal/domain/entity"
)

type IPInfoRepository interface {
	FindAll(context.Context) ([]entity.IPInfoDto, error)
	Insert(context.Context, entity.IPInfoDto) (int, error)
	FindIdByIP(ctx context.Context, ip string) (int, error)
}

type IPInfoUsecase struct {
	repository IPInfoRepository
}

type IPInfoSerivce interface {
	Create(context.Context, entity.IPInfoDto) (int, error)
	GetAll(context.Context) ([]entity.IPInfoDto, error)
	GetByIp(ctx context.Context, ip string) (int, error)
}

func NewIPInfoService(repository IPInfoRepository) IPInfoSerivce {
	return &IPInfoUsecase{repository: repository}
}

func (i *IPInfoUsecase) Create(ctx context.Context, info entity.IPInfoDto) (int, error) {
	return i.repository.Insert(ctx, info)
}

func (i *IPInfoUsecase) GetAll(ctx context.Context) ([]entity.IPInfoDto, error) {
	return i.repository.FindAll(ctx)
}

func (s *IPInfoUsecase) GetByIp(ctx context.Context, ip string) (int, error) {
	return s.repository.FindIdByIP(ctx, ip)
}
