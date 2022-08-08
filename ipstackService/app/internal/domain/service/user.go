package service

import (
	"context"
	"ipstack/internal/domain/entity"
)

type UserSerivce interface {
	Create(context.Context, entity.User) (int, error)
	GetAll(context.Context) ([]entity.User, error)
	GetByNickname(ctx context.Context, nickname string) (int, error)
}

type UserRepository interface {
	FindAll(context.Context) ([]entity.User, error)
	Insert(context.Context, entity.User) (int, error)
	FindIdByNickname(ctx context.Context, nickname string) (int, error)
}

type UserUsecase struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) UserSerivce {
	return &UserUsecase{repository: repository}
}

func (s *UserUsecase) Create(ctx context.Context, user entity.User) (int, error) {
	return s.repository.Insert(ctx, user)
}

func (s *UserUsecase) GetAll(ctx context.Context) ([]entity.User, error) {
	return s.repository.FindAll(ctx)
}

func (s *UserUsecase) GetByNickname(ctx context.Context, nickname string) (int, error) {
	return s.repository.FindIdByNickname(ctx, nickname)
}
