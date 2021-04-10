package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/message-broker/internal/service"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type nameController struct {
	logger   *zap.Logger
	queue    *service.Queue
	validate *validator.Validate
}

type namerRequest struct {
	Value string `validate:"required"`
}

func NameController(logger *zap.Logger, queue *service.Queue, validate *validator.Validate) Controller {
	return &nameController{logger: logger, queue: queue, validate: validate}
}

func (c *nameController) SetRoutes(gin *gin.Engine) {
	gin.GET("/name", c.GetAction)
	gin.PUT("/name", c.PutAction)
}

func (c *nameController) GetAction(ctx *gin.Context) {
	timeout := ctx.Query("timeout")
	if timeout != "" {
		c.dequeueWithTimeout(ctx, timeout)
		return
	}

	value := c.queue.Dequeue()
	if value != "" {
		ctx.JSON(http.StatusOK, successHttpResponse{Value: value})
		return
	}

	ctx.JSON(http.StatusNotFound, &errorHttpResponse{ErrorMessage: notFoundErrorMessage})
}

func (c *nameController) PutAction(ctx *gin.Context) {
	value := ctx.Query("v")
	request := &namerRequest{Value: value}
	err := c.validate.Struct(request)
	if err != nil {
		c.logger.Error("validation error", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
	}

	c.queue.Enqueue(value)
	ctx.JSON(http.StatusOK, nil)
}

func (c *nameController) dequeueWithTimeout(ctx *gin.Context, timeout string) {
	c.logger.Debug("dequeue with timeout")
	t, err := strconv.Atoi(timeout)
	if err != nil {
		c.logger.Error("timeout should be an integer value", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
		return
	}

	value, err := c.queue.DequeueWithTimeout(t)
	if err == service.ErrorQueueTimeoutLimit {
		c.logger.Error(service.ErrorQueueTimeoutLimit.Error())
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
		return
	}

	if value != "" {
		ctx.JSON(http.StatusOK, successHttpResponse{Value: value})
		return
	}

	ctx.JSON(http.StatusNotFound, &errorHttpResponse{ErrorMessage: notFoundErrorMessage})
}
