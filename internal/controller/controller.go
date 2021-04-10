package controller

import (
	"github.com/gin-gonic/gin"
)

type Controller interface {
	SetRoutes(gin *gin.Engine)
}

const (
	badRequestErrorMessage = "Bad Request"
	notFoundErrorMessage   = "Not Found"
)
