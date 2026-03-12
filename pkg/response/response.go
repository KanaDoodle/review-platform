package response

import "github.com/gin-gonic/gin"

type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Resp{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(200, Resp{
		Code:    code,
		Message: message,
	})
}