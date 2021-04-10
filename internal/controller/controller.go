package controller

import "github.com/gin-gonic/gin"

type Controller interface {
	SetRoutes(gin *gin.Engine)
}

type errorHttpResponse struct {
	ErrorMessage string `json:"error_message"`
}

type successHttpResponse struct {
	Value string `json:"value"`
}

const badRequestErrorMessage = "Bad Request"
const notFoundErrorMessage = "Not Found"
