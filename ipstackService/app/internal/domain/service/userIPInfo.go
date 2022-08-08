package service

import (
	"context"
	"ipstack/internal/domain/entity"
	"ipstack/pkg/logging"

	"github.com/jackc/pgx"
)

type UserIPInfoSerivce interface {
	Create(ctx context.Context, nickname string, ip entity.IPInfoDto) error
	GetAll(context.Context) ([]entity.UserIPInfo, error)
	GetAllUserIpInfo(ctx context.Context, nickname string) ([]entity.IPInfoDto, error)
}

type UserIPInfoRepository interface {
	FindAll(context.Context) ([]entity.UserIPInfo, error)
	Insert(context.Context, entity.UserIPInfo) error
	FindUserIpAddresess(ctx context.Context, userId int) ([]entity.IPInfoDto, error)
}

type UserIPInfoUsecase struct {
	userService    UserSerivce
	ipInfoService  IPInfoSerivce
	userIPInfoRepo UserIPInfoRepository
	logger         *logging.Logger
}

func NewUserIPInfoService(uIpRepo UserIPInfoRepository, uServ UserSerivce, ipServ IPInfoSerivce) UserIPInfoSerivce {
	return &UserIPInfoUsecase{userIPInfoRepo: uIpRepo, userService: uServ, ipInfoService: ipServ}
}

/*
	not user:
		not ip:
			isert user
			insert ip
			insert user ip
		isert user
		insert user ip
	else
		not ip:
			insert ip
			insert user ip
		insert user ip


*/
func (s *UserIPInfoUsecase) Create(ctx context.Context, nickname string, ipInfo entity.IPInfoDto) error {
	userId, err := s.userService.GetByNickname(ctx, nickname)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			userId, err = s.userService.Create(ctx, entity.User{Nickname: nickname})
			if err != nil {
				s.logger.Errorf("can't save user due to %v ", err)
				return err
			}
		} else {
			return err
		}
	}
	ipId, err := s.ipInfoService.GetByIp(ctx, ipInfo.IP)
	if err != nil {
		if err.Error() == pgx.ErrNoRows.Error() {
			ipId, err = s.ipInfoService.Create(ctx, ipInfo)
			if err != nil {
				s.logger.Errorf("can't save user's ip-info due to %v ", err)
				return err
			}
		} else {
			return err
		}
	}

	return s.userIPInfoRepo.Insert(ctx, entity.UserIPInfo{UserId: userId, IPId: ipId})

}

func (s *UserIPInfoUsecase) GetAll(ctx context.Context) ([]entity.UserIPInfo, error) {
	return s.userIPInfoRepo.FindAll(ctx)
}

func (s *UserIPInfoUsecase) GetAllUserIpInfo(ctx context.Context, nickname string) ([]entity.IPInfoDto, error) {
	userId, err := s.userService.GetByNickname(ctx, nickname)
	if err != nil {
		s.logger.Errorf("cannot find user's info due to %v", err)
		return nil, err
	}
	return s.userIPInfoRepo.FindUserIpAddresess(ctx, userId)
}
