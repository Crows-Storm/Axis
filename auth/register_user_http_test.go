package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRegisterUserHandler is a mock implementation of RegisterUserCommandHandler
type MockRegisterUserHandler struct {
	mock.Mock
}

func (m *MockRegisterUserHandler) Handle(ctx context.Context, cmd command.RegisterUserCommand) (struct{}, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(struct{}), args.Error(1)
}

// setupTestRouter creates a test router with the HTTPServer
func setupTestRouter(mockHandler *MockRegisterUserHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create application with mock handler
	application := app.Application{
		Commands: app.Commands{
			RegisterUser: mockHandler,
		},
	}

	httpServer := HTTPServer{app: application}

	// Register the route
	router.POST("/api/register", httpServer.Register)

	return router
}

func TestRegister_Success(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	requestBody := command.RegisterUserCommand{
		LoginId:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	// Mock expects the handler to be called with correct parameters
	mockHandler.On("Handle", mock.Anything, requestBody).Return(struct{}{}, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "successfully", response["data"])
	mockHandler.AssertExpectations(t)
}

func TestRegister_InvalidJSON(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	invalidJSON := `{"loginId": "testuser", "password": "Test@1234", "email": }`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["message"].(string), "Invalid request body")
	mockHandler.AssertNotCalled(t, "Handle")
}

func TestRegister_MissingRequiredFields(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	testCases := []struct {
		name        string
		requestBody interface{}
		description string
	}{
		{
			name: "missing_loginId",
			requestBody: map[string]string{
				"password": "Test@1234",
				"email":    "test@example.com",
			},
			description: "LoginId is required",
		},
		{
			name: "missing_password",
			requestBody: map[string]string{
				"loginId": "testuser",
				"email":   "test@example.com",
			},
			description: "Password is required",
		},
		{
			name: "missing_email",
			requestBody: map[string]string{
				"loginId":  "testuser",
				"password": "Test@1234",
			},
			description: "Email is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Act
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestRegister_EmptyBody(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	mockHandler.AssertNotCalled(t, "Handle")
}

func TestRegister_HandlerError(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	requestBody := command.RegisterUserCommand{
		LoginId:  "existinguser",
		Password: "Test@1234",
		Email:    "existing@example.com",
	}

	// Mock expects the handler to return an error
	expectedError := errors.New("user already exists")
	mockHandler.On("Handle", mock.Anything, requestBody).Return(struct{}{}, expectedError)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // server.ErrorWithCode 返回 200 (需要改进)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["message"].(string), "Faield")
	assert.Contains(t, response["message"].(string), "user already exists")
	mockHandler.AssertExpectations(t)
}

func TestRegister_WrongContentType(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	requestBody := command.RegisterUserCommand{
		LoginId:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "text/plain") // 错误的 Content-Type
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["message"].(string), "Invalid request body")
}

func TestRegister_LargePayload(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	// 创建一个超大的 payload
	largePassword := string(make([]byte, 10000)) // 10KB password
	requestBody := command.RegisterUserCommand{
		LoginId:  "testuser",
		Password: largePassword,
		Email:    "test@example.com",
	}

	mockHandler.On("Handle", mock.Anything, requestBody).Return(struct{}{}, nil)

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// 应该考虑添加 payload 大小限制
}

func TestRegister_ConcurrentRequests(t *testing.T) {
	// Arrange
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	requestBody := command.RegisterUserCommand{
		LoginId:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	mockHandler.On("Handle", mock.Anything, requestBody).Return(struct{}{}, nil).Times(10)

	// Act - 并发发送 10 个请求
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			body, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Assert
	mockHandler.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkRegister_Success(b *testing.B) {
	mockHandler := new(MockRegisterUserHandler)
	router := setupTestRouter(mockHandler)

	requestBody := command.RegisterUserCommand{
		LoginId:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	mockHandler.On("Handle", mock.Anything, requestBody).Return(struct{}{}, nil)

	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
