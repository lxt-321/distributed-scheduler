package common

import "github.com/gin-gonic/gin"

// Response 统一 HTTP 响应结构（XXL-Job 风格）
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// 返回码
const (
	CodeSuccess = 200
	CodeFail    = 500
)

func JSON(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(200, Response{Code: code, Msg: msg, Data: data})
}

func Success(c *gin.Context, data interface{}) {
	JSON(c, CodeSuccess, "success", data)
}

func Fail(c *gin.Context, msg string) {
	JSON(c, CodeFail, msg, nil)
}
