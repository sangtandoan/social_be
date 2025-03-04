package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sangtandoan/social/internal/models/dto"
	"github.com/sangtandoan/social/internal/store"
	"github.com/sangtandoan/social/internal/utils"
)

type CreatePostPayload struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(c *gin.Context) error {
	payload := &CreatePostPayload{}
	if err := utils.ReadJSON(c, payload); err != nil {
		return err
	}

	post := &store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
		UserID:  1,
	}

	err := app.store.Posts.Create(c.Request.Context(), post)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, post)
	return nil
}

func (a *application) getPostHandler(c *gin.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	post, err := a.store.Posts.GetByID(c.Request.Context(), int64(id))
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, post)
	return nil
}

func (a *application) getPostsHandler(c *gin.Context) error {
	data, err := a.store.Posts.GetAll(c.Request.Context())
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, data)
	return nil
}

func (a *application) getUserFeedHandler(c *gin.Context) {
	userID, ok := c.Request.Context().Value("userID").(int64)
	if !ok {
		c.Error(fmt.Errorf("userID not found in context"))
		return
	}

	var req dto.UserFeedRequest
	req.Offset = 0
	req.Limit = 10

	err := c.ShouldBindQuery(&req)
	if err != nil {
		c.Error(err)
		return
	}

	req.ID = userID

	res, err := a.store.Posts.GetUserFeed(c.Request.Context(), &req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, utils.NewApiResponse("Fetch feed successfully", res))
}

func (a *application) updatePostHandler(c *gin.Context) error {
	var req store.UpdatePostParams

	// Only for req body
	err := c.ShouldBind(&req)
	if err != nil {
		return utils.ErrInvalidJSON
	}
	id, err := strconv.Atoi(c.Param("id"))
	req.ID = int64(id)

	if err != nil {
		return err
	}

	post, err := a.store.Posts.UpdatePost(c.Request.Context(), &req)
	if err != nil {
		return err
	}

	c.JSON(http.StatusOK, post)
	return nil
}
