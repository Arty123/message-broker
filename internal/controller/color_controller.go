package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/message-broker/internal/service"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type colorController struct {
	logger   *zap.Logger
	queue    *service.Queue
	validate *validator.Validate
}

type colorRequest struct {
	Value string `validate:"required"`
}

func ColorController(logger *zap.Logger, queue *service.Queue, validate *validator.Validate) Controller {
	return &colorController{logger: logger, queue: queue, validate: validate}
}

func (c *colorController) SetRoutes(gin *gin.Engine) {
	gin.GET("/color", c.GetAction)
	gin.PUT("/color", c.PutAction)
}

func (c *colorController) GetAction(ctx *gin.Context) {
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

func (c *colorController) PutAction(ctx *gin.Context) {
	value := ctx.Query("v")
	request := &colorRequest{Value: value}
	err := c.validate.Struct(request)
	if err != nil {
		c.logger.Error("validation error", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
	}

	c.queue.Enqueue(value)
}

func (c *colorController) dequeueWithTimeout(ctx *gin.Context, timeout string) {
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
