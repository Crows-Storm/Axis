package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGinContextFallbackCarriesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.ContextWithFallback = true
	r.Use(RequestContextMiddleware("auth-1"))
	r.POST("/api/register", func(c *gin.Context) {
		assert.Equal(t, "6485-9574-3e5c", contextx.FromContext(c).RequestID())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/register", nil)
	req.Header.Set(contextx.HeaderRequestID, "6485-9574-3e5c")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestContextMiddleware_RequiresRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContextMiddleware("user-local-1"))
	r.GET("/api/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body Response[any]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, CodeBadRequest, body.Code)
}

func TestRequestContextMiddleware_BuildsContextAndResponseHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestContextMiddleware("user-local-1"))
	r.GET("/ok", func(c *gin.Context) {
		reader := contextx.FromGinContext(c)
		assert.Equal(t, "req-1", reader.RequestID())
		assert.NotEmpty(t, reader.CorrelationID())
		assert.NotEqual(t, "forged-corr", reader.CorrelationID())
		assert.NotEmpty(t, reader.CausationID())
		assert.NotEqual(t, "forged-cause", reader.CausationID())
		assert.Equal(t, "user-local-1", reader.InstanceID())
		assert.NotEqual(t, "upstream-span", reader.SpanID())
		assert.Equal(t, "upstream-span", reader.ParentSpanID())
		assert.Equal(t, "Bearer abc", reader.Authorization())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.Header.Set(contextx.HeaderRequestID, "req-1")
	req.Header.Set("X-Correlation-ID", "forged-corr")
	req.Header.Set("X-Causation-ID", "forged-cause")
	req.Header.Set(contextx.HeaderSpanID, "upstream-span")
	req.Header.Set(contextx.HeaderAuthorization, "Bearer abc")
	req.Header.Set("X-Instance-ID", "forged")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "req-1", w.Header().Get(contextx.HeaderRequestID))
	assert.Equal(t, "req-1", w.Header().Get(contextx.HeaderTraceID))
	assert.Empty(t, w.Header().Get("X-Correlation-ID"))
	assert.Empty(t, w.Header().Get("X-Causation-ID"))
	assert.NotEmpty(t, w.Header().Get(contextx.HeaderSpanID))
	assert.NotEqual(t, "upstream-span", w.Header().Get(contextx.HeaderSpanID))
	assert.Empty(t, w.Header().Get("X-Instance-ID"))
}
