package command

import (
	"context"
	"errors"
	"testing"

	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const passwordHash = "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetInfo(id int64) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) GetByLoginId(ctx context.Context, loginId string) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) GetPasswordByLoginId(ctx context.Context, loginId string) string {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) CreateBatch(ctx context.Context, users []*domain.User) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) Disable(ctx context.Context, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) SoftDelete(ctx context.Context, userId int64) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockUserRepository) ExistsWithTransaction(ctx context.Context, id int64, loginId string, email string) (bool, error) {
	result := m.Called(ctx, id, loginId, email)
	return result.Bool(0), result.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	result := m.Called(ctx, user)
	return result.Get(0).(*domain.User), result.Error(1)
}

type MockMetricsClient struct {
	mock.Mock
}

func (m *MockMetricsClient) Inc(key string, value int) {
	m.Called(key, value)
}

func (m *MockMetricsClient) IncCounter(name string, tags ...string) {
	m.Called(name, tags)
}

func (m *MockMetricsClient) RecordTimer(name string, duration float64, tags ...string) {
	m.Called(name, duration, tags)
}

func (m *MockMetricsClient) RecordHistogram(name string, value float64, tags ...string) {
	m.Called(name, value, tags)
}

// ==================== Handler 集成测试 ====================
func TestCreateUserCommandHandler(t *testing.T) {
	tests := []struct {
		name       string
		cmd        CreateUserCommand
		setupMock  func(repo *MockUserRepository, mc *MockMetricsClient)
		wantResult any
		wantErr    string
	}{
		{
			name: "成功创建用户",
			cmd: CreateUserCommand{
				LoginID:  "testuser",
				Password: "SecurePass123",
				Email:    "test@example.com",
			},
			setupMock: func(repo *MockUserRepository, mc *MockMetricsClient) {
				repo.On("ExistsWithTransaction",
					mock.Anything, mock.AnythingOfType("int64"), mock.Anything, mock.Anything,
				).Return(false, nil).Once()

				repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
					Return(&domain.User{}, nil).Once()

				mc.On("Inc", mock.Anything, mock.AnythingOfType("int")).Maybe()
			},
			wantResult: struct{}{},
		},
		{
			name: "用户已存在返回错误",
			cmd: CreateUserCommand{
				LoginID:  "existinguser",
				Password: "SecurePass123",
				Email:    "existing@example.com",
			},
			setupMock: func(repo *MockUserRepository, mc *MockMetricsClient) {
				repo.On("ExistsWithTransaction",
					mock.Anything, mock.AnythingOfType("int64"), mock.Anything, mock.Anything,
				).Return(true, nil).Once()

				mc.On("Inc", mock.Anything, mock.AnythingOfType("int")).Maybe()
			},
			wantErr: "user already exists",
		},
		{
			name: "数据库错误",
			cmd: CreateUserCommand{
				LoginID:  "testuser",
				Password: "SecurePass123",
				Email:    "test@example.com",
			},
			setupMock: func(repo *MockUserRepository, mc *MockMetricsClient) {
				repo.On("ExistsWithTransaction",
					mock.Anything, mock.AnythingOfType("int64"), mock.Anything, mock.Anything,
				).Return(false, nil).Once()

				repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
					Return(nil, errors.New("database connection failed")).Once()

				mc.On("Inc", mock.Anything, mock.AnythingOfType("int")).Maybe()
			},
			wantErr: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockUserRepository)
			mc := new(MockMetricsClient)

			tt.setupMock(repo, mc)

			handler := NewCreateUserCommandHandler(repo, mc)

			result, err := handler.Handle(context.Background(), tt.cmd)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Equal(t, struct{}{}, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}

			repo.AssertExpectations(t)
			mc.AssertExpectations(t)
		})
	}
}

func TestCreateUserCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     CreateUserCommand
		wantErr string
	}{
		{
			name: "合法命令",
			cmd: CreateUserCommand{
				LoginID:  "testuser",
				Password: "SecurePass123",
				Email:    "test@example.com",
			},
		},
		{
			name:    "LoginID为空",
			cmd:     CreateUserCommand{LoginID: "", Password: passwordHash, Email: "test@example.com"},
			wantErr: "invalid params",
		},
		{
			name:    "Password为空",
			cmd:     CreateUserCommand{LoginID: "testuser", Password: "", Email: "test@example.com"},
			wantErr: "invalid params",
		},
		{
			name:    "Email为空",
			cmd:     CreateUserCommand{LoginID: "testuser", Password: passwordHash, Email: ""},
			wantErr: "invalid params",
		},
		{
			name:    "所有字段为空",
			cmd:     CreateUserCommand{},
			wantErr: "invalid params",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ==================== Benchmark ====================
func BenchmarkCreateUserCommandHandler(b *testing.B) {
	repo := new(MockUserRepository)
	mc := new(MockMetricsClient)

	repo.On("ExistsWithTransaction",
		mock.Anything, mock.AnythingOfType("int64"), mock.Anything, mock.Anything,
	).Return(false, nil)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(&domain.User{}, nil)

	mc.On("Inc", mock.Anything, mock.Anything).Maybe()
	mc.On("IncCounter", mock.Anything, mock.Anything).Maybe()
	mc.On("RecordTimer", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mc.On("RecordHistogram", mock.Anything, mock.Anything, mock.Anything).Maybe()

	handler := NewCreateUserCommandHandler(repo, mc)
	cmd := CreateUserCommand{
		LoginID:  "testuser",
		Password: "SecurePass123",
		Email:    "test@example.com",
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.Handle(ctx, cmd)
	}
}
