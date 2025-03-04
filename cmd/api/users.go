package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sangtandoan/social/internal/models/dto"
	"github.com/sangtandoan/social/internal/utils"
)

func (a *application) createUserHandler(c *gin.Context) {
	var req dto.CreateUserRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.Error(err)
		return
	}

	err = utils.Validator.Struct(&req)
	if err != nil {
		c.Error(err)
		return
	}
}
