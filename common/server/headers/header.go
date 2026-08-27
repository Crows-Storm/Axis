package headers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ==================== 基础获取方法 ====================

// Get 获取请求头值
func Get(c *gin.Context, key string) string {
	return c.GetHeader(key)
}

// GetOrDefault 获取请求头值，不存在则返回默认值
func GetOrDefault(c *gin.Context, key, defaultValue string) string {
	if value := c.GetHeader(key); value != "" {
		return value
	}
	return defaultValue
}

// Has 检查请求头是否存在
func Has(c *gin.Context, key string) bool {
	return c.GetHeader(key) != ""
}

// ==================== 常见请求头获取 ====================

// GetAuthorization 获取 Authorization 头
func GetAuthorization(c *gin.Context) string {
	return c.GetHeader("Authorization")
}

// GetBearerToken 从 Authorization 头提取 Bearer Token
func GetBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// GetContentType 获取 Content-Type
func GetContentType(c *gin.Context) string {
	return c.GetHeader("Content-Type")
}

// IsJSONContent 检查是否为 JSON 请求
func IsJSONContent(c *gin.Context) bool {
	contentType := GetContentType(c)
	return strings.Contains(contentType, "application/json")
}

// IsFormContent 检查是否为表单请求
func IsFormContent(c *gin.Context) bool {
	contentType := GetContentType(c)
	return strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data")
}

// GetAccept 获取 Accept 头
func GetAccept(c *gin.Context) string {
	return c.GetHeader("Accept")
}

// GetAcceptLanguage 获取 Accept-Language
func GetAcceptLanguage(c *gin.Context) string {
	return c.GetHeader("Accept-Language")
}

// GetAcceptEncoding 获取 Accept-Encoding
func GetAcceptEncoding(c *gin.Context) string {
	return c.GetHeader("Accept-Encoding")
}

// ==================== 用户代理和客户端信息 ====================

// GetUserAgent 获取 User-Agent
func GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// GetReferer 获取 Referer
func GetReferer(c *gin.Context) string {
	return c.GetHeader("Referer")
}

// GetOrigin 获取 Origin
func GetOrigin(c *gin.Context) string {
	return c.GetHeader("Origin")
}

// GetHost 获取 Host
func GetHost(c *gin.Context) string {
	return c.Request.Host
}

// GetClientIP 获取客户端 IP（考虑代理）
func GetClientIP(c *gin.Context) string {
	// 1. 检查 X-Forwarded-For
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 2. 检查 X-Real-IP
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 3. 使用 gin 的方法
	return c.ClientIP()
}

// GetRealIP 获取真实 IP（别名）
func GetRealIP(c *gin.Context) string {
	return GetClientIP(c)
}

// ==================== 安全相关 ====================

// GetCSRFToken 获取 CSRF Token
func GetCSRFToken(c *gin.Context) string {
	if token := c.GetHeader("X-CSRF-Token"); token != "" {
		return token
	}
	return c.GetHeader("X-XSRF-Token")
}

// GetAPIKey 获取 API Key
func GetAPIKey(c *gin.Context) string {
	if key := c.GetHeader("X-API-Key"); key != "" {
		return key
	}
	return GetBearerToken(c)
}

// GetSignature 获取签名
func GetSignature(c *gin.Context) string {
	return c.GetHeader("X-Signature")
}

// GetTimestamp 获取时间戳
func GetTimestamp(c *gin.Context) int64 {
	timestamp := c.GetHeader("X-Timestamp")
	if timestamp == "" {
		return 0
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return 0
	}
	return ts
}

// GetNonce 获取随机数（防重放攻击）
func GetNonce(c *gin.Context) string {
	return c.GetHeader("X-Nonce")
}

// ==================== 请求追踪 ====================

// GetRequestID 获取请求 ID
func GetRequestID(c *gin.Context) string {
	if id := c.GetHeader("X-Request-ID"); id != "" {
		return id
	}
	if id := c.GetHeader("X-Correlation-ID"); id != "" {
		return id
	}
	if id := c.GetHeader("X-Trace-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}

// GetCorrelationID 获取关联 ID
func GetCorrelationID(c *gin.Context) string {
	if id := c.GetHeader("X-Correlation-ID"); id != "" {
		return id
	}
	return GetRequestID(c)
}

// GetTraceID 获取追踪 ID
func GetTraceID(c *gin.Context) string {
	if id := c.GetHeader("X-Trace-ID"); id != "" {
		return id
	}
	return GetRequestID(c)
}

// GetSpanID 获取 Span ID
func GetSpanID(c *gin.Context) string {
	return c.GetHeader("X-Span-ID")
}

// ==================== 设备信息 ====================

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID       string `json:"device_id"`
	DeviceType     string `json:"device_type"`
	DeviceName     string `json:"device_name"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	Browser        string `json:"browser"`
	BrowserVersion string `json:"browser_version"`
}

// GetDeviceInfo 获取设备信息
func GetDeviceInfo(c *gin.Context) *DeviceInfo {
	return &DeviceInfo{
		DeviceID:       c.GetHeader("X-Device-ID"),
		DeviceType:     c.GetHeader("X-Device-Type"),
		DeviceName:     c.GetHeader("X-Device-Name"),
		OS:             c.GetHeader("X-OS"),
		OSVersion:      c.GetHeader("X-OS-Version"),
		Browser:        c.GetHeader("X-Browser"),
		BrowserVersion: c.GetHeader("X-Browser-Version"),
	}
}

// ==================== 分页信息 ====================

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// GetPagination 从请求头获取分页信息
func GetPagination(c *gin.Context) *PaginationInfo {
	page := 1
	if p := c.GetHeader("X-Page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	pageSize := 20
	if ps := c.GetHeader("X-Page-Size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 && val <= 100 {
			pageSize = val
		}
	}

	return &PaginationInfo{
		Page:     page,
		PageSize: pageSize,
	}
}

// ==================== 国际化 ====================

// GetLocale 获取语言区域
func GetLocale(c *gin.Context) string {
	if lang := c.GetHeader("Accept-Language"); lang != "" {
		parts := strings.Split(lang, ",")
		if len(parts) > 0 {
			langPart := strings.Split(parts[0], ";")[0]
			return strings.TrimSpace(langPart)
		}
	}

	if locale := c.GetHeader("X-Locale"); locale != "" {
		return locale
	}

	return "en-US"
}

// GetLanguage 获取语言（简写）
func GetLanguage(c *gin.Context) string {
	locale := GetLocale(c)
	parts := strings.Split(locale, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return "en"
}

// ==================== 缓存相关 ====================

// GetCacheControl 获取 Cache-Control
func GetCacheControl(c *gin.Context) string {
	return c.GetHeader("Cache-Control")
}

// GetIfModifiedSince 获取 If-Modified-Since
func GetIfModifiedSince(c *gin.Context) time.Time {
	val := c.GetHeader("If-Modified-Since")
	if val == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC1123, val)
	if err != nil {
		return time.Time{}
	}
	return t
}

// GetIfNoneMatch 获取 If-None-Match (ETag)
func GetIfNoneMatch(c *gin.Context) string {
	return c.GetHeader("If-None-Match")
}

// ==================== 压缩相关 ====================

// GetCompression 获取支持的压缩算法
func GetCompression(c *gin.Context) []string {
	encoding := c.GetHeader("Accept-Encoding")
	if encoding == "" {
		return []string{}
	}
	parts := strings.Split(encoding, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

// SupportGzip 是否支持 Gzip
func SupportGzip(c *gin.Context) bool {
	encodings := GetCompression(c)
	for _, e := range encodings {
		if strings.Contains(e, "gzip") {
			return true
		}
	}
	return false
}

// ==================== 请求头对象 ====================

// Headers 完整的请求头对象
type Headers struct {
	Authorization  string            `json:"authorization"`
	ContentType    string            `json:"content_type"`
	Accept         string            `json:"accept"`
	AcceptLanguage string            `json:"accept_language"`
	UserAgent      string            `json:"user_agent"`
	Referer        string            `json:"referer"`
	Origin         string            `json:"origin"`
	Host           string            `json:"host"`
	ClientIP       string            `json:"client_ip"`
	RequestID      string            `json:"request_id"`
	CSRFToken      string            `json:"csrf_token"`
	DeviceID       string            `json:"device_id"`
	All            map[string]string `json:"all"`
}

// GetAllHeaders 获取所有请求头
func GetAllHeaders(c *gin.Context) *Headers {
	allHeaders := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			allHeaders[key] = values[0]
		}
	}

	return &Headers{
		Authorization:  GetAuthorization(c),
		ContentType:    GetContentType(c),
		Accept:         GetAccept(c),
		AcceptLanguage: GetAcceptLanguage(c),
		UserAgent:      GetUserAgent(c),
		Referer:        GetReferer(c),
		Origin:         GetOrigin(c),
		Host:           GetHost(c),
		ClientIP:       GetClientIP(c),
		RequestID:      GetRequestID(c),
		CSRFToken:      GetCSRFToken(c),
		DeviceID:       c.GetHeader("X-Device-ID"),
		All:            allHeaders,
	}
}
