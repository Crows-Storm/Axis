package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response[D any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    D      `json:"data,omitempty"`
}

type PageData[D any] struct {
	List     D     `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

const (
	CodeSuccess      = 0
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// Success 成功响应
func Success[D any](c *gin.Context, data D) {
	c.JSON(http.StatusOK, Response[D]{
		Code:    CodeSuccess,
		Message: "Successfully",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage[D any](c *gin.Context, message string, data D) {
	c.JSON(http.StatusOK, Response[D]{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应（自动推断 HTTP 状态码）
func Error(c *gin.Context, code int, message string) {
	httpStatus := 200
	if code >= 400 && code < 600 {
		httpStatus = code
	}
	c.JSON(httpStatus, Response[any]{
		Code:    code,
		Message: message,
	})
}

// ErrorWithStatus 错误响应（自定义 HTTP 状态码）
func ErrorWithStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response[any]{
		Code:    code,
		Message: message,
	})
}

// Page 分页响应
func Page[D any](c *gin.Context, list D, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response[PageData[D]]{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData[D]{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}
