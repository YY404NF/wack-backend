package httpserver

import "github.com/gin-gonic/gin"

type responseBody[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func ok[T any](c *gin.Context, data T) {
	c.JSON(200, responseBody[T]{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func fail(c *gin.Context, status int, message string) {
	c.JSON(status, responseBody[gin.H]{
		Code:    status,
		Message: message,
		Data:    gin.H{},
	})
}
