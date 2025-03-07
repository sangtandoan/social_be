package repo

import (
	"context"

	"github.com/sangtandoan/practice/internal/models"
)

type UserRepo interface {
	GetUsers(ctx context.Context) ([]*models.User, error)
}

type userRepo struct{}

func NewUserRepo() *userRepo {
	return &userRepo{}
}

func (repo *userRepo) GetUsers(ctx context.Context) ([]*models.User, error) {
	var data []*models.User

	user := &models.User{ID: 1, Username: "sang", Email: "sang@gmail.com", Age: 24}
	data = append(data, user)

	user = &models.User{ID: 2, Username: "bin", Email: "bin@gmail.com", Age: 17}
	data = append(data, user)

	return data, nil
}
