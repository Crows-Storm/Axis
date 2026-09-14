package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Crows-Storm/Axis/common/genproto/userpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockUserServiceClient is a mock implementation of userpb.UserServiceClient
type MockUserServiceClient struct {
	mock.Mock
}

func (m *MockUserServiceClient) CreateUser(ctx context.Context, in *userpb.CreateUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m *MockUserServiceClient) GetUserById(ctx context.Context, in *userpb.GetUserByIdRequest, opts ...grpc.CallOption) (*userpb.GetUserByIdResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.GetUserByIdResponse), args.Error(1)
}

func (m *MockUserServiceClient) GetUserByLoginID(ctx context.Context, in *userpb.GetUserByLoginIDRequest, opts ...grpc.CallOption) (*userpb.GetUserByLoginIDResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.GetUserByLoginIDResponse), args.Error(1)
}

// TestNewUserGRPC tests the constructor
func TestNewUserGRPC(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)

	// Act
	userGRPC := NewUserGRPC(mockClient)

	// Assert
	assert.NotNil(t, userGRPC)
	assert.Equal(t, mockClient, userGRPC.client)
}

// TestCreateUser_Success tests successful user creation
func TestCreateUser_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	expectedResp := &emptypb.Empty{}
	mockClient.On("CreateUser", ctx, req).Return(expectedResp, nil)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	mockClient.AssertExpectations(t)
}

// TestCreateUser_DuplicateUser tests creating a duplicate user
func TestCreateUser_DuplicateUser(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "existinguser",
		Password: "Test@1234",
		Email:    "existing@example.com",
	}

	expectedErr := status.Error(codes.AlreadyExists, "user already exists")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
	assert.Contains(t, err.Error(), "user already exists")
	mockClient.AssertExpectations(t)
}

// TestCreateUser_InvalidRequest tests creating user with invalid data
func TestCreateUser_InvalidRequest(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()

	testCases := []struct {
		name        string
		req         *userpb.CreateUserRequest
		expectedErr error
		description string
	}{
		{
			name: "empty_LoginID",
			req: &userpb.CreateUserRequest{
				LoginID:  "",
				Password: "Test@1234",
				Email:    "test@example.com",
			},
			expectedErr: status.Error(codes.InvalidArgument, "LoginID is required"),
			description: "LoginID cannot be empty",
		},
		{
			name: "empty_password",
			req: &userpb.CreateUserRequest{
				LoginID:  "testuser",
				Password: "",
				Email:    "test@example.com",
			},
			expectedErr: status.Error(codes.InvalidArgument, "password is required"),
			description: "Password cannot be empty",
		},
		{
			name: "invalid_email",
			req: &userpb.CreateUserRequest{
				LoginID:  "testuser",
				Password: "Test@1234",
				Email:    "invalid-email",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid email format"),
			description: "Email format must be valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.On("CreateUser", ctx, tc.req).Return((*emptypb.Empty)(nil), tc.expectedErr)

			// Act
			resp, err := userGRPC.CreateUser(ctx, tc.req)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Equal(t, status.Code(tc.expectedErr), status.Code(err))
			mockClient.AssertExpectations(t)
		})
	}
}

// TestCreateUser_ContextTimeout tests context timeout
func TestCreateUser_ContextTimeout(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	expectedErr := status.Error(codes.DeadlineExceeded, "context deadline exceeded")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestCreateUser_NetworkError tests network failure
func TestCreateUser_NetworkError(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	expectedErr := status.Error(codes.Unavailable, "service unavailable")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Unavailable, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestGetUserById_Success tests successful retrieval by ID
func TestGetUserById_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByIdRequest{
		ID: 1234253231,
	}

	expectedResp := &userpb.GetUserByIdResponse{
		ID:      int64(1234253231),
		LoginID: "testuser",
		Email:   "test@example.com",
		Status:  1,
	}
	mockClient.On("GetUserById", ctx, req).Return(expectedResp, nil)

	// Act
	resp, err := userGRPC.GetUserById(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	assert.Equal(t, int64(1234253231), resp.ID)
	assert.Equal(t, "testuser", resp.LoginID)
	assert.Equal(t, "test@example.com", resp.Email)
	mockClient.AssertExpectations(t)
}

// TestGetUserById_NotFound tests user not found scenario
func TestGetUserById_NotFound(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByIdRequest{
		ID: 1234253232,
	}

	expectedErr := status.Error(codes.NotFound, "user not found")
	mockClient.On("GetUserById", ctx, req).Return((*userpb.GetUserByIdResponse)(nil), expectedErr)

	// Act
	resp, err := userGRPC.GetUserById(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestGetUserById_EmptyId tests empty ID validation
func TestGetUserById_EmptyId(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByIdRequest{
		ID: 0,
	}

	expectedErr := status.Error(codes.InvalidArgument, "id is required")
	mockClient.On("GetUserById", ctx, req).Return((*userpb.GetUserByIdResponse)(nil), expectedErr)

	// Act
	resp, err := userGRPC.GetUserById(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestGetUserByLoginID_Success tests successful retrieval by login ID
func TestGetUserByLoginID_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByLoginIDRequest{
		LoginID: "testuser",
	}

	expectedResp := &userpb.GetUserByLoginIDResponse{
		ID:      int64(1234253233),
		LoginID: "testuser",
		Email:   "test@example.com",
		Status:  1,
	}
	mockClient.On("GetUserByLoginID", ctx, req).Return(expectedResp, nil)

	// Act
	resp, err := userGRPC.GetUserByLoginID(ctx, req)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedResp, resp)
	assert.Equal(t, "testuser", resp.LoginID)
	mockClient.AssertExpectations(t)
}

// TestGetUserByLoginID_NotFound tests user not found by login ID
func TestGetUserByLoginID_NotFound(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByLoginIDRequest{
		LoginID: "nonexistentuser",
	}

	expectedErr := status.Error(codes.NotFound, "user not found")
	mockClient.On("GetUserByLoginID", ctx, req).Return((*userpb.GetUserByLoginIDResponse)(nil), expectedErr)

	// Act
	resp, err := userGRPC.GetUserByLoginID(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.NotFound, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestGetUserByLoginID_EmptyLoginID tests empty login ID validation
func TestGetUserByLoginID_EmptyLoginID(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByLoginIDRequest{
		LoginID: "",
	}

	expectedErr := status.Error(codes.InvalidArgument, "LoginID is required")
	mockClient.On("GetUserByLoginID", ctx, req).Return((*userpb.GetUserByLoginIDResponse)(nil), expectedErr)

	// Act
	resp, err := userGRPC.GetUserByLoginID(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestUserGRPC_NilClient tests behavior with nil client (should panic in production)
func TestUserGRPC_NilClient(t *testing.T) {
	// Arrange & Act
	userGRPC := NewUserGRPC(nil)

	// Assert - should handle nil gracefully or panic appropriately
	assert.NotNil(t, userGRPC)
	// Note: 实际调用会 panic，应该在构造函数中添加 nil 检查
}

// TestUserGRPC_ContextCancellation tests context cancellation
func TestUserGRPC_ContextCancellation(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	expectedErr := status.Error(codes.Canceled, "context canceled")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Canceled, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestUserGRPC_ConcurrentCalls tests concurrent gRPC calls
func TestUserGRPC_ConcurrentCalls(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	numCalls := 10

	// Setup expectations for concurrent calls
	for i := 0; i < numCalls; i++ {
		req := &userpb.GetUserByIdRequest{ID: int64(1234253235)}
		resp := &userpb.GetUserByIdResponse{
			ID:      int64(1234253235),
			LoginID: "testuser",
			Email:   "test@example.com",
		}
		mockClient.On("GetUserById", ctx, req).Return(resp, nil).Once()
	}

	// Act - Make concurrent calls
	done := make(chan bool, numCalls)
	for i := 0; i < numCalls; i++ {
		go func() {
			req := &userpb.GetUserByIdRequest{ID: int64(1234253235)}
			resp, err := userGRPC.GetUserById(ctx, req)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < numCalls; i++ {
		<-done
	}

	// Assert
	mockClient.AssertExpectations(t)
}

// TestUserGRPC_RetryLogic tests retry behavior (if implemented)
func TestUserGRPC_TransientError(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	// Simulate transient error
	expectedErr := status.Error(codes.Unavailable, "temporary unavailable")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Unavailable, status.Code(err))
	// Note: 如果实现了重试逻辑，应该验证重试次数
	mockClient.AssertExpectations(t)
}

// TestUserGRPC_InternalServerError tests internal server error handling
func TestUserGRPC_InternalServerError(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	expectedErr := status.Error(codes.Internal, "internal server error")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.Internal, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestUserGRPC_PermissionDenied tests permission denied scenario
func TestUserGRPC_PermissionDenied(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByIdRequest{
		ID: int64(1234253236),
	}

	expectedErr := status.Error(codes.PermissionDenied, "permission denied")
	mockClient.On("GetUserById", ctx, req).Return((*userpb.GetUserByIdResponse)(nil), expectedErr)

	// Act
	resp, err := userGRPC.GetUserById(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
	mockClient.AssertExpectations(t)
}

// TestUserGRPC_UnknownError tests unknown/generic error
func TestUserGRPC_UnknownError(t *testing.T) {
	// Arrange
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	expectedErr := errors.New("unknown error")
	mockClient.On("CreateUser", ctx, req).Return((*emptypb.Empty)(nil), expectedErr)

	// Act
	resp, err := userGRPC.CreateUser(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockClient.AssertExpectations(t)
}

// Benchmark tests
func BenchmarkCreateUser(b *testing.B) {
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.CreateUserRequest{
		LoginID:  "testuser",
		Password: "Test@1234",
		Email:    "test@example.com",
	}

	mockClient.On("CreateUser", ctx, req).Return(&emptypb.Empty{}, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = userGRPC.CreateUser(ctx, req)
	}
}

func BenchmarkGetUserById(b *testing.B) {
	mockClient := new(MockUserServiceClient)
	userGRPC := NewUserGRPC(mockClient)

	ctx := context.Background()
	req := &userpb.GetUserByIdRequest{ID: int64(1234253237)}
	resp := &userpb.GetUserByIdResponse{
		ID:      int64(1234253237),
		LoginID: "testuser",
		Email:   "test@example.com",
	}

	mockClient.On("GetUserById", ctx, req).Return(resp, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = userGRPC.GetUserById(ctx, req)
	}
}
