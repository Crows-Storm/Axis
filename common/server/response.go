package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Response[D any] struct {
	Code      Code      `json:"code"`
	Message   string    `json:"message"`
	Data      D         `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type PageData[D any] struct {
	List     D     `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

func Success[D any](c *gin.Context, data D) {
	c.AbortWithStatusJSON(http.StatusOK, Response[D]{
		Code:      CodeSuccess,
		Message:   "Successfully",
		Data:      data,
		Timestamp: time.Now(),
	})
}

func SuccessWithCode[D any](c *gin.Context, code Code, data D) {
	c.AbortWithStatusJSON(http.StatusOK, Response[D]{
		Code:      code,
		Message:   code.String(),
		Data:      data,
		Timestamp: time.Now(),
	})
}

func SuccessWithMessage[D any](c *gin.Context, message string, data D) {
	c.AbortWithStatusJSON(http.StatusOK, Response[D]{
		Code:      CodeSuccess,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func Error(c *gin.Context, message string) {
	httpStatus := http.StatusOK
	// The HTTP status code returned is 200, but the status code inside the response body is the business result status code
	//if code >= 400 && code < 600 {
	//	httpStatus = code
	//}
	c.AbortWithStatusJSON(httpStatus, Response[any]{
		Code:      CodeBusinessError,
		Message:   message,
		Timestamp: time.Now(),
	})
}

func ErrorWithCode(c *gin.Context, code Code) {
	httpStatus := http.StatusOK
	// The HTTP status code returned is 200, but the status code inside the response body is the business result status code
	//if code >= 400 && code < 600 {
	//	httpStatus = code
	//}
	c.AbortWithStatusJSON(httpStatus, Response[any]{
		Code:      code,
		Message:   code.String(),
		Timestamp: time.Now(),
	})
}

func ErrorWithHttpCode(c *gin.Context, code Code, httpStatus int) {
	// The HTTP status code returned is 200, but the status code inside the response body is the business result status code
	//if code >= 400 && code < 600 {
	//	httpStatus = code
	//}
	c.AbortWithStatusJSON(httpStatus, Response[any]{
		Code:      code,
		Message:   code.String(),
		Timestamp: time.Now(),
	})
}

func ErrorWithCodeAndMessage(c *gin.Context, code Code, msg string) {
	// The HTTP status code returned is 200, but the status code inside the response body is the business result status code
	//if code >= 400 && code < 600 {
	//	httpStatus = code
	//}
	httpStatus := http.StatusOK
	c.AbortWithStatusJSON(httpStatus, Response[any]{
		Code:      code,
		Message:   msg,
		Timestamp: time.Now(),
	})
}

func ErrorWithStatusUnauthorized(c *gin.Context) {
	httpStatus := http.StatusUnauthorized
	// The HTTP status code returned is 200, but the status code inside the response body is the business result status code
	//if code >= 400 && code < 600 {
	//	httpStatus = code
	//}
	c.AbortWithStatusJSON(httpStatus, Response[any]{
		Code:      CodeUnauthorized,
		Message:   CodeUnauthorized.String(),
		Timestamp: time.Now(),
	})
}

func ErrorWithStatus(c *gin.Context, httpStatus int, code Code, message string) {
	c.AbortWithStatusJSON(httpStatus, Response[any]{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	})
}

func AbortWithStatus(c *gin.Context, code Code, statusCode int) {
	c.AbortWithStatusJSON(statusCode,
		Response[any]{
			Code:      code,
			Message:   code.String(),
			Timestamp: time.Now(),
		})
}

func Page[D any](c *gin.Context, list D, total int64, page, pageSize int) {
	c.AbortWithStatusJSON(http.StatusOK, Response[PageData[D]]{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData[D]{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
		Timestamp: time.Now(),
	})
}

type Code int

const (
	// CodeSuccess (0-99)
	CodeSuccess  Code = 0
	CodeCreated  Code = 1
	CodeAccepted Code = 2

	// CodeBadRequest (1000-1999)
	CodeBadRequest           Code = 1000
	CodeValidationFailed     Code = 1001
	CodeInvalidParams        Code = 1002
	CodeMissingParams        Code = 1003
	CodeUnsupportedMediaType Code = 1004
	CodeTooManyRequests      Code = 1005
	CodeRequestTimeout       Code = 1006
	CodeInvalidRange         Code = 1007

	// CodeUnauthorized (2000-2999)
	CodeUnauthorized            Code = 2000
	CodeInvalidToken            Code = 2001
	CodeTokenExpired            Code = 2002
	CodeTokenMissing            Code = 2003
	CodeForbidden               Code = 2004
	CodeInsufficientPermissions Code = 2005
	CodeAccountLocked           Code = 2006
	CodeAccountDisabled         Code = 2007
	CodePasswordExpired         Code = 2008
	CodeInvalidCredentials      Code = 2009

	// CodeNotFound (3000-3999)
	CodeNotFound         Code = 3000
	CodeResourceNotFound Code = 3001
	CodeUserNotFound     Code = 3002
	CodeProductNotFound  Code = 3003
	CodeOrderNotFound    Code = 3004
	CodeResourceConflict Code = 3005
	CodeDuplicateEntry   Code = 3006
	CodeResourceLocked   Code = 3007
	CodeResourceDeleted  Code = 3008

	// CodeBusinessError (4000-4999)
	CodeBusinessError         Code = 4000
	CodeInsufficientBalance   Code = 4001
	CodeOrderStatusInvalid    Code = 4002
	CodePaymentFailed         Code = 4003
	CodeInventoryInsufficient Code = 4004
	CodeOperationNotAllowed   Code = 4005
	CodeVerificationFailed    Code = 4006
	CodeRateLimitExceeded     Code = 4007
	CodeServiceUnavailable    Code = 4008

	// CodeInternalServerError (5000-5999)
	CodeInternalServerError Code = 5000
	CodeDatabaseError       Code = 5001
	CodeCacheError          Code = 5002
	CodeThirdPartyError     Code = 5003
	CodeFileOperationError  Code = 5004
	CodeNetworkError        Code = 5005
	CodeTimeoutError        Code = 5006
	CodeSerializationError  Code = 5007
	CodeDependencyError     Code = 5008

	// CodeThirdPartyServiceError (6000-6999)
	CodeThirdPartyServiceError Code = 6000
	CodeWeChatError            Code = 6001
	CodeAlipayError            Code = 6002
	CodeSMSError               Code = 6003
	CodeEmailError             Code = 6004
	CodeOAuthError             Code = 6005
	CodePaymentGatewayError    Code = 6006
	CodeStorageServiceError    Code = 6007
)

var codeMessages = map[Code]string{
	CodeSuccess:  "success",
	CodeCreated:  "resource created successfully",
	CodeAccepted: "request accepted",

	CodeBadRequest:           "bad request",
	CodeValidationFailed:     "validation failed",
	CodeInvalidParams:        "invalid parameters",
	CodeMissingParams:        "missing required parameters",
	CodeUnsupportedMediaType: "unsupported media type",
	CodeTooManyRequests:      "too many requests",
	CodeRequestTimeout:       "request timeout",
	CodeInvalidRange:         "invalid range",

	CodeUnauthorized:            "unauthorized",
	CodeInvalidToken:            "invalid token",
	CodeTokenExpired:            "token expired",
	CodeTokenMissing:            "token missing",
	CodeForbidden:               "forbidden",
	CodeInsufficientPermissions: "insufficient permissions",
	CodeAccountLocked:           "account locked",
	CodeAccountDisabled:         "account disabled",
	CodePasswordExpired:         "password expired",
	CodeInvalidCredentials:      "invalid credentials",

	CodeNotFound:         "resource not found",
	CodeResourceNotFound: "resource not found",
	CodeUserNotFound:     "user not found",
	CodeProductNotFound:  "product not found",
	CodeOrderNotFound:    "order not found",
	CodeResourceConflict: "resource conflict",
	CodeDuplicateEntry:   "duplicate entry",
	CodeResourceLocked:   "resource locked",
	CodeResourceDeleted:  "resource deleted",

	CodeBusinessError:         "business error",
	CodeInsufficientBalance:   "insufficient balance",
	CodeOrderStatusInvalid:    "invalid order status",
	CodePaymentFailed:         "payment failed",
	CodeInventoryInsufficient: "insufficient inventory",
	CodeOperationNotAllowed:   "operation not allowed",
	CodeVerificationFailed:    "verification failed",
	CodeRateLimitExceeded:     "rate limit exceeded",
	CodeServiceUnavailable:    "service unavailable",

	CodeInternalServerError: "internal server error",
	CodeDatabaseError:       "database error",
	CodeCacheError:          "cache error",
	CodeThirdPartyError:     "third party service error",
	CodeFileOperationError:  "file operation error",
	CodeNetworkError:        "network error",
	CodeTimeoutError:        "timeout error",
	CodeSerializationError:  "serialization error",
	CodeDependencyError:     "dependency error",

	CodeThirdPartyServiceError: "third party service error",
	CodeWeChatError:            "wechat service error",
	CodeAlipayError:            "alipay service error",
	CodeSMSError:               "sms service error",
	CodeEmailError:             "email service error",
	CodeOAuthError:             "oauth service error",
	CodePaymentGatewayError:    "payment gateway error",
	CodeStorageServiceError:    "storage service error",
}

func (c Code) String() string {
	if msg, ok := codeMessages[c]; ok {
		return msg
	}
	return "unknown error"
}

func (c Code) Message() string {
	return c.String()
}

func (c Code) Int() int {
	return int(c)
}

// IsSuccess returns true if the code indicates a successful operation (0-99).
func (c Code) IsSuccess() bool {
	return c >= 0 && c < 100
}

// IsClientError returns true if the code indicates a client-side error (1000-1999).
func (c Code) IsClientError() bool {
	return c >= 1000 && c < 2000
}

// IsAuthError returns true if the code indicates an authentication or authorization error (2000-2999).
func (c Code) IsAuthError() bool {
	return c >= 2000 && c < 3000
}

// IsResourceError returns true if the code indicates a resource-related error (3000-3999).
func (c Code) IsResourceError() bool {
	return c >= 3000 && c < 4000
}

// IsBusinessError returns true if the code indicates a business logic error (4000-4999).
func (c Code) IsBusinessError() bool {
	return c >= 4000 && c < 5000
}

// IsServerError returns true if the code indicates a server-side error (5000-5999).
func (c Code) IsServerError() bool {
	return c >= 5000 && c < 6000
}

// IsThirdPartyError returns true if the code indicates a third-party service error (6000-6999).
func (c Code) IsThirdPartyError() bool {
	return c >= 6000 && c < 7000
}

// IsError returns true if the code indicates any error (non-zero code).
func (c Code) IsError() bool {
	return c != CodeSuccess
}
