package service

import (
	"context"

	"github.com/sangtandoan/practice/internal/models"
	"github.com/sangtandoan/practice/internal/repo"
)

type UserService interface {
	GetUsers(ctx context.Context) ([]*models.User, error)
}

type userService struct {
	userRepo repo.UserRepo
}

func NewUserService(userRepo repo.UserRepo) *userService {
	return &userService{userRepo}
}

func (s *userService) GetUsers(ctx context.Context) ([]*models.User, error) {
	data, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
