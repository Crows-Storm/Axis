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

func Success[D any](c *gin.Context, data D) {
	c.JSON(http.StatusOK, Response[D]{
		Code:    CodeSuccess,
		Message: "Successfully",
		Data:    data,
	})
}

func SuccessWithMessage[D any](c *gin.Context, message string, data D) {
	c.JSON(http.StatusOK, Response[D]{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	httpStatus := 200
	// The HTTP status code returned is 200, but the status code inside the response body is the business result status code
	//if code >= 400 && code < 600 {
	//	httpStatus = code
	//}
	c.JSON(httpStatus, Response[any]{
		Code:    code,
		Message: message,
	})
}

func ErrorWithStatus(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Response[any]{
		Code:    code,
		Message: message,
	})
}

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
