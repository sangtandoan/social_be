package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangtandoan/social/internal/models/dto"
	"github.com/sangtandoan/social/internal/models/params"
	"github.com/sangtandoan/social/internal/service"
	"github.com/sangtandoan/social/internal/utils"
	"golang.org/x/crypto/bcrypt"
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

	var res dto.CreateUserResponse

	err = a.store.Tx.WithTx(c.Request.Context(), func(txCtx context.Context) error {
		_, err := a.store.Users.GetByEmail(txCtx, req.Email)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		req.Password = string(hashedPassword)
		user, err := a.store.Users.Create(txCtx, &req)
		if err != nil {
			return err
		}

		token := uuid.New().String()
		hash := sha256.Sum256([]byte(token))
		hashedToken := hex.EncodeToString(hash[:])

		params := params.CreateInvitationParams{
			UserID:    user.ID,
			Token:     hashedToken,
			ExpiresAt: time.Now().Add(10 * time.Minute),
		}

		err = a.store.Invitations.CreateInvitation(txCtx, &params)
		if err != nil {
			return err
		}

		res.UserID = user.ID
		res.Email = user.Email
		res.Username = user.Username
		res.Token = token
		return nil
	})
	if err != nil {
		c.Error(err)
		return
	}

	emailReq := service.SendRequest{
		To: []string{"sanghutao143@gmail.com"},
		Data: &service.ConfirmData{
			Username: res.Username,
			Token:    res.Token,
		},
		Temp: service.ConfirmTemplate,
	}

	if err := a.mailer.SendWithRetry(&emailReq, 3); err != nil {
		// implement saga pattern
		if err := a.store.Users.Delete(c.Request.Context(), res.UserID); err != nil {
			c.Error(err)
		}
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, utils.NewApiResponse("created user successfully", res))
}

func (a *application) activateUserHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.Error(utils.NewApiError(http.StatusNotFound, "activate token not found"))
		return
	}

	err := a.store.Tx.WithTx(c.Request.Context(), func(txCtx context.Context) error {
		userID, err := a.store.Invitations.GetUserIDFromInvitation(txCtx, token)
		if err != nil {
			return err
		}

		err = a.store.Users.Activate(txCtx, userID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, utils.NewApiResponse("activate successfully", nil))
}

func (a *application) loginHandler(c *gin.Context) {
	var req dto.LoginRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.Error(utils.ErrInvalidJSON)
		return
	}

	user, err := a.store.Users.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.Error(utils.ErrNotFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.Error(utils.ErrUnauthorized)
		return
	}

	// TODO: return access and refresh tokens

	c.JSON(http.StatusOK, utils.NewApiResponse("login successfully", nil))
}
