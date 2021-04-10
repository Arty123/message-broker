package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/message-broker/internal/service"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type errorHttpResponse struct {
	ErrorMessage string `json:"error_message"`
}

type successHttpResponse struct {
	Value string `json:"value"`
}

type queueController struct {
	logger        *zap.Logger
	queueResolver service.QueueResolver
	validate      *validator.Validate
}

type getRequest struct {
	Name string `validate:"required,oneof=name color"`
}

type putRequest struct {
	Name  string `validate:"required,oneof=name color"`
	Value string `validate:"required"`
}

func QueueController(logger *zap.Logger, queueResolver service.QueueResolver, validate *validator.Validate) Controller {
	return &queueController{logger: logger, queueResolver: queueResolver, validate: validate}
}

func (c *queueController) SetRoutes(gin *gin.Engine) {
	gin.GET("/:queue", c.GetAction)
	gin.PUT("/:queue", c.PutAction)
}

func (c *queueController) GetAction(ctx *gin.Context) {
	queueName := ctx.Param("queue")
	request := &getRequest{Name: queueName}
	err := c.validate.Struct(request)
	if err != nil {
		c.logger.Error("validation error", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
		return
	}

	queue := c.queueResolver.ResolveQueue(queueName)

	timeout := ctx.Query("timeout")
	if timeout != "" {
		c.dequeueWithTimeout(ctx, timeout, queue)
		return
	}

	value := queue.Dequeue()
	if value != "" {
		ctx.JSON(http.StatusOK, successHttpResponse{Value: value})
		return
	}

	ctx.JSON(http.StatusNotFound, &errorHttpResponse{ErrorMessage: notFoundErrorMessage})
}

func (c *queueController) PutAction(ctx *gin.Context) {
	value := ctx.Query("v")
	queueName := ctx.Param("queue")

	request := &putRequest{Name: queueName, Value: value}
	err := c.validate.Struct(request)
	if err != nil {
		c.logger.Error("validation error", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
		return
	}

	queue := c.queueResolver.ResolveQueue(queueName)
	queue.Enqueue(value)

	ctx.JSON(http.StatusOK, "OK")
}

func (c *queueController) dequeueWithTimeout(ctx *gin.Context, timeout string, queue *service.Queue) {
	c.logger.Debug("dequeue with timeout")
	t, err := strconv.Atoi(timeout)
	if err != nil {
		c.logger.Error("timeout should be an integer value", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, &errorHttpResponse{ErrorMessage: badRequestErrorMessage})
		return
	}

	value, err := queue.DequeueWithTimeout(t)
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
