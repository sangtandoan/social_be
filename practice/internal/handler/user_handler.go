package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sangtandoan/practice/internal/models/dto"
	"github.com/sangtandoan/practice/internal/pkg/apperrors"
	"github.com/sangtandoan/practice/internal/pkg/response"
	"github.com/sangtandoan/practice/internal/pkg/validator"
	"github.com/sangtandoan/practice/internal/service"
)

type UserHandler interface {
	GetUsers(c *gin.Context) error
	CreateUser(c *gin.Context) error
}

type userHandler struct {
	s         service.UserService
	validator *validator.CustomValidator
}

func NewUserHandler(s service.UserService, validator *validator.CustomValidator) *userHandler {
	return &userHandler{s, validator}
}

func (h *userHandler) GetUsers(c *gin.Context) error {
	data, err := h.s.GetUsers(c.Request.Context())
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, response.NewApiResponse("Get all users successfully!", data))
	return nil
}

func (h *userHandler) CreateUser(c *gin.Context) error {
	var req dto.CreateUserRequest

	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		c.Error(apperrors.ErrInvalidJSON)
		return err
	}

	err = h.validator.Validate(req)
	if err != nil {
		c.Error(err)
		return err
	}

	return nil
}
