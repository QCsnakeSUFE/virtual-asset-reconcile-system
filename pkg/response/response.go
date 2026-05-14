package response

import (
	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	TraceID string `json:"trace_id,omitempty"`
}

func Success(c *gin.Context, data any) {
	traceID, _ := c.Get("trace_id")
	c.JSON(200, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		TraceID: traceID.(string),
	})
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	// 若失败，则返回错误消息 message、错误码 code 以及 http 状态 httpStatus
	traceID, _ := c.Get("trace_id")
	c.JSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
		TraceID: traceID.(string),
	})
}
