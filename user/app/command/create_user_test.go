package command

import (
	"context"
	"errors"
	"testing"

	"github.com/Crows-Storm/Axis/common/decorator"
	decoratormock "github.com/Crows-Storm/Axis/common/decorator/mocks"
	commuser "github.com/Crows-Storm/Axis/common/domain/user"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	usermock "github.com/Crows-Storm/Axis/user/domain/user/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const passwordHash = "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92"

func TestCreateUserCommandHandler(t *testing.T) {
	tests := []struct {
		name      string
		cmd       CreateUserCommand
		setupMock func(repo *usermock.MockRepository, mc *decoratormock.MockMetricsClient)
		wantErr   error
	}{
		{
			name: "Created Success",
			cmd: CreateUserCommand{
				LoginID:  "testuser",
				Password: passwordHash,
				Email:    "test@example.com",
			},
			setupMock: func(repo *usermock.MockRepository, mc *decoratormock.MockMetricsClient) {
				repo.On("ExistsWithTransaction", context.Background(), mock.AnythingOfType("uint64"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false, nil)

				repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
					Return(&domain.User{}, nil)

				mc.On("Inc", mock.Anything, mock.Anything).Maybe()
			},
		},
		{
			name: "user already exists Return an error",
			cmd: CreateUserCommand{
				LoginID:  "existinguser",
				Password: passwordHash,
				Email:    "existing@example.com",
			},
			setupMock: func(repo *usermock.MockRepository, mc *decoratormock.MockMetricsClient) {
				repo.On("ExistsWithTransaction", context.Background(), mock.AnythingOfType("uint64"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(true, nil)

				// cannot set expectation, because ExistsWithTransaction check return true
				//repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
				//	Return(nil, errors.New("user already exists"))
				mc.On("Inc", mock.Anything, mock.Anything).Maybe()
			},
			wantErr: commuser.ErrUserAlreadyExists,
		},
		{
			name: "Database Error",
			cmd: CreateUserCommand{
				LoginID:  "testuser",
				Password: passwordHash,
				Email:    "test@example.com",
			},
			setupMock: func(repo *usermock.MockRepository, mc *decoratormock.MockMetricsClient) {
				repo.On("ExistsWithTransaction", context.Background(), mock.AnythingOfType("uint64"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(false, nil)

				repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil, errors.New("database connection failed"))

				mc.On("Inc", mock.Anything, mock.Anything).Maybe()
			},
			wantErr: errors.New("database connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(usermock.MockRepository)
			mc := new(decoratormock.MockMetricsClient)
			tt.setupMock(repo, mc)

			handler := NewCreateUserCommandHandler(repo, mc)

			// Act
			result, err := handler.Handle(context.Background(), tt.cmd)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantErr.Error())
				assert.Equal(t, struct{}{}, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, struct{}{}, result)
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
			name: "Correct command",
			cmd: CreateUserCommand{
				LoginID:  "testuser",
				Password: passwordHash,
				Email:    "test@example.com",
			},
		},
		{
			name:    "LoginID is empty",
			cmd:     CreateUserCommand{LoginID: "", Password: passwordHash, Email: "test@example.com"},
			wantErr: decorator.InvalidCommand.Error(),
		},
		{
			name:    "Password is empty",
			cmd:     CreateUserCommand{LoginID: "testuser", Password: "", Email: "test@example.com"},
			wantErr: decorator.InvalidCommand.Error(),
		},
		{
			name:    "Email is empty",
			cmd:     CreateUserCommand{LoginID: "testuser", Password: passwordHash, Email: ""},
			wantErr: decorator.InvalidCommand.Error(),
		},
		{
			name:    "All fields are empty",
			cmd:     CreateUserCommand{},
			wantErr: decorator.InvalidCommand.Error(),
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

// command: go test -bench=BenchmarkCreateUserCommandHandler -count=5
// output:
//
// goos: darwin
// goarch: arm64
// pkg: github.com/Crows-Storm/Axis/user/app/command
// cpu: Apple M5
// BenchmarkCreateUserCommandHandler-10               59046             20150 ns/op
// BenchmarkCreateUserCommandHandler-10               59733             20339 ns/op
// BenchmarkCreateUserCommandHandler-10               59510             30101 ns/op
// BenchmarkCreateUserCommandHandler-10               57523             20652 ns/op
// BenchmarkCreateUserCommandHandler-10               58363             20541 ns/op
// PASS
// ok      github.com/Crows-Storm/Axis/user/app/command    9.004s
func BenchmarkCreateUserCommandHandler(b *testing.B) {
	repo := new(usermock.MockRepository)
	mc := new(decoratormock.MockMetricsClient)

	repo.On("ExistsWithTransaction", context.Background(), mock.AnythingOfType("uint64"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(true, nil)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(&domain.User{}, nil)
	mc.On("Inc", mock.Anything, mock.Anything).Maybe()

	handler := NewCreateUserCommandHandler(repo, mc)
	cmd := CreateUserCommand{
		LoginID:  "testuser",
		Password: passwordHash,
		Email:    "test@example.com",
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.Handle(ctx, cmd)
	}
}
