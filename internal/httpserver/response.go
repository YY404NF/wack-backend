package httpserver

import "github.com/gin-gonic/gin"

type responseBody struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ok(c *gin.Context, data interface{}) {
	c.JSON(200, responseBody{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func fail(c *gin.Context, status int, message string) {
	c.JSON(status, responseBody{
		Code:    status,
		Message: message,
		Data:    gin.H{},
	})
}
