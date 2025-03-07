package handler

import (
	"github.com/sangtandoan/practice/internal/pkg/validator"
	"github.com/sangtandoan/practice/internal/service"
)

type Handler struct {
	User UserHandler
}

func NewHanlder(s *service.Service, validator *validator.CustomValidator) *Handler {
	return &Handler{
		User: NewUserHandler(s.User, validator),
	}
}
