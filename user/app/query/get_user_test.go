package query

//import (
//	"context"
//	"errors"
//	"testing"
//
//	"github.com/Crows-Storm/Axis/common/decorator"
//	"github.com/Crows-Storm/Axis/common/metrics"
//	domain "github.com/Crows-Storm/Axis/user/domain/user"
//	"github.com/golang/mock/gomock"
//	"github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestGetUserQueryHandler_Handle(t *testing.T) {
//	ctx := context.Background()
//
//	tests := []struct {
//		name          string
//		query         GetUserQuery
//		mockSetup     func(mockRepo *MockRepository)
//		expectedUser  *domain.User
//		expectedError error
//	}{
//		{
//			name:  "成功获取用户信息",
//			query: GetUserQuery{Id: 123},
//			mockSetup: func(mockRepo *MockRepository) {
//				mockRepo.EXPECT().
//					GetInfo(int64(123)).
//					Return(&domain.User{
//						ID:    123,
//						Name:  "张三",
//						Email: "zhangsan@example.com",
//					}, nil).
//					Times(1)
//			},
//			expectedUser: &domain.User{
//				ID:    123,
//				Name:  "张三",
//				Email: "zhangsan@example.com",
//			},
//			expectedError: nil,
//		},
//		{
//			name:  "用户不存在",
//			query: GetUserQuery{Id: 999},
//			mockSetup: func(mockRepo *MockRepository) {
//				mockRepo.EXPECT().
//					GetInfo(int64(999)).
//					Return(nil, domain.ErrUserNotFound).
//					Times(1)
//			},
//			expectedUser:  nil,
//			expectedError: domain.ErrUserNotFound,
//		},
//		{
//			name:  "数据库错误",
//			query: GetUserQuery{Id: 456},
//			mockSetup: func(mockRepo *MockRepository) {
//				mockRepo.EXPECT().
//					GetInfo(int64(456)).
//					Return(nil, errors.New("database connection failed")).
//					Times(1)
//			},
//			expectedUser:  nil,
//			expectedError: errors.New("database connection failed"),
//		},
//		{
//			name:  "ID 为 0 的边界情况",
//			query: GetUserQuery{Id: 0},
//			mockSetup: func(mockRepo *MockRepository) {
//				mockRepo.EXPECT().
//					GetInfo(int64(0)).
//					Return(nil, domain.ErrInvalidUserID).
//					Times(1)
//			},
//			expectedUser:  nil,
//			expectedError: domain.ErrInvalidUserID,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// 创建 mock 控制器
//			ctrl := gomock.NewController(t)
//			defer ctrl.Finish()
//
//			// 创建 mock repository
//			mockRepo := NewMockRepository(ctrl)
//
//			// 设置 mock 预期行为
//			if tt.mockSetup != nil {
//				tt.mockSetup(mockRepo)
//			}
//
//			// 创建 handler（使用真实的 logger 和 metrics）
//			logger := logrus.NewEntry(logrus.New())
//			// 注意：如果 decorator 需要 metrics 接口，这里也需要 mock
//			// 但这里我们假设 decorator 不依赖具体的 metrics 实现
//			mockMetrics := &metrics.TodoMetrics{}
//
//			handler := NewGetUserQueryHandler(mockRepo, logger, mockMetrics)
//
//			// 执行测试
//			user, err := handler.Handle(ctx, tt.query)
//
//			// 断言结果
//			if tt.expectedError != nil {
//				assert.Error(t, err)
//				assert.Equal(t, tt.expectedError.Error(), err.Error())
//				assert.Nil(t, user)
//			} else {
//				assert.NoError(t, err)
//				assert.Equal(t, tt.expectedUser, user)
//			}
//		})
//	}
//}
//
//// 2. 如果需要测试 decorator 的装饰器功能，可以单独测试
//func TestGetUserQueryHandler_WithDecorators(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockRepo := NewMockRepository(ctrl)
//
//	// 设置 mock 返回
//	expectedUser := &domain.User{ID: 1, Name: "测试用户"}
//	mockRepo.EXPECT().
//		GetInfo(int64(1)).
//		Return(expectedUser, nil).
//		Times(1)
//
//	logger := logrus.NewEntry(logrus.New())
//	mockMetrics := &MockMetricsClient{}
//
//	handler := NewGetUserQueryHandler(mockRepo, logger, mockMetrics)
//
//	ctx := context.Background()
//	query := GetUserQuery{Id: 1}
//
//	user, err := handler.Handle(ctx, query)
//
//	assert.NoError(t, err)
//	assert.Equal(t, expectedUser, user)
//}
//
//// 3. 测试构造函数
//func TestNewGetUserQueryHandler(t *testing.T) {
//	t.Run("正常创建", func(t *testing.T) {
//		ctrl := gomock.NewController(t)
//		defer ctrl.Finish()
//
//		mockRepo := NewMockRepository(ctrl)
//		logger := logrus.NewEntry(logrus.New())
//		mockMetrics := &MockMetricsClient{}
//
//		// 不应该 panic
//		handler := NewGetUserQueryHandler(mockRepo, logger, mockMetrics)
//		assert.NotNil(t, handler)
//	})
//
//	t.Run("repo 为 nil 时 panic", func(t *testing.T) {
//		logger := logrus.NewEntry(logrus.New())
//		mockMetrics := &MockMetricsClient{}
//
//		// 应该 panic
//		assert.PanicsWithValue(t, "nil User Repository", func() {
//			NewGetUserQueryHandler(nil, logger, mockMetrics)
//		})
//	})
//}
//
//// 4. 如果需要，添加基准测试
//func BenchmarkGetUserQueryHandler_Handle(b *testing.B) {
//	ctrl := gomock.NewController(b)
//	defer ctrl.Finish()
//
//	mockRepo := NewMockRepository(ctrl)
//	mockRepo.EXPECT().
//		GetInfo(gomock.Any()).
//		Return(&domain.User{ID: 1, Name: "测试"}, nil).
//		AnyTimes()
//
//	logger := logrus.NewEntry(logrus.New())
//	mockMetrics := &MockMetricsClient{}
//	handler := NewGetUserQueryHandler(mockRepo, logger, mockMetrics)
//
//	ctx := context.Background()
//	query := GetUserQuery{Id: 1}
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		_, _ = handler.Handle(ctx, query)
//	}
//}
