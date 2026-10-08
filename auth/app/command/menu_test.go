package command

import (
	"context"
	"errors"
	"testing"

	domain "github.com/Crows-Storm/Axis/auth/domain/menu"
	menumock "github.com/Crows-Storm/Axis/auth/domain/menu/mocks"
	decoratormock "github.com/Crows-Storm/Axis/common/decorator/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateMenuCommandHandler(t *testing.T) {
	tests := []struct {
		name      string
		cmd       CreateMenuCommand
		setupMock func(repo *menumock.MockRepository, mc *decoratormock.MockMetricsClient, bus *decoratormock.MockEventBus)
		wantErr   error
	}{
		{
			name: "Create Root Menu Success",
			cmd: CreateMenuCommand{ // stored db id = 11
				ParentID: 0, // is root
				Code:     "user",
				Name:     "User Management",
				Path:     "/user",
				Icon:     "user",
				Sort:     10,
			},
			setupMock: func(repo *menumock.MockRepository, mc *decoratormock.MockMetricsClient, bus *decoratormock.MockEventBus) {
				// mock Expectations
				repo.On("Create", mock.Anything, mock.AnythingOfType("*menu.Menu")).Return(nil) // error is empty
				mc.On("Inc", mock.Anything, mock.Anything)
				bus.On("Publish", context.Background(), mock.Anything).Return(nil)
			},
			wantErr: nil,
		}, {
			name: "Create Sub Menu Success",
			cmd: CreateMenuCommand{
				ParentID: 11, // is sub, Assume the ID of the last added menu is 11
				Code:     "deleted user page",
				Name:     "User Deleted Management",
				Path:     "/user/deleted", // status=deleted the user list
				Icon:     "user",
				Sort:     10,
			},
			setupMock: func(repo *menumock.MockRepository, mc *decoratormock.MockMetricsClient, bus *decoratormock.MockEventBus) {
				// mock Expectations
				repo.On("Create", mock.Anything, mock.AnythingOfType("*menu.Menu")).Return(nil) // error is empty
				mc.On("Inc", mock.Anything, mock.Anything)
				bus.On("Publish", context.Background(), mock.Anything).Return(nil)
			},
			wantErr: nil,
		}, {
			name: "Creation Root menu failed due to missing Name",
			cmd: CreateMenuCommand{
				ParentID: 11, // is sub, Assume the ID of the last added menu is 11
				Code:     "deleted user page",
				//Name:     "User Deleted Management",
				Path: "/user/deleted", // status=deleted the user list
				Icon: "user",
				Sort: 10,
			},
			setupMock: func(repo *menumock.MockRepository, mc *decoratormock.MockMetricsClient, bus *decoratormock.MockEventBus) {
				mc.On("Inc", mock.Anything, mock.Anything)
			},
			wantErr: errors.New("code and name are required"),
		}, {
			name: "Creation Root menu failed due to missing Code",
			cmd: CreateMenuCommand{
				ParentID: 11, // is sub, Assume the ID of the last added menu is 11
				//Code:     "deleted user page",
				Name: "User Deleted Management",
				Path: "/user/deleted", // status=deleted the user list
				Icon: "user",
				Sort: 10,
			},
			setupMock: func(repo *menumock.MockRepository, mc *decoratormock.MockMetricsClient, bus *decoratormock.MockEventBus) {
				mc.On("Inc", mock.Anything, mock.Anything)
			},
			wantErr: errors.New("code and name are required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(menumock.MockRepository)
			mc := new(decoratormock.MockMetricsClient)
			bus := new(decoratormock.MockEventBus)

			tt.setupMock(repo, mc, bus)

			handler := NewCreateMenuCommandHandler(repo, mc, bus)

			// Act
			result, err := handler.Handle(context.Background(), tt.cmd)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantErr.Error())
				assert.IsType(t, &domain.Menu{}, result)
			} else {
				require.NoError(t, err)
				assert.IsType(t, &domain.Menu{}, result)

				assert.True(t, result.ParentID == tt.cmd.ParentID)
				assert.True(t, result.Code == tt.cmd.Code)
				assert.True(t, result.Name == tt.cmd.Name)
				assert.True(t, result.Path == tt.cmd.Path)
				assert.True(t, result.Icon == tt.cmd.Icon)
				assert.True(t, result.Sort == tt.cmd.Sort)
			}

			repo.AssertExpectations(t)
			mc.AssertExpectations(t)
			bus.AssertExpectations(t)
		})
	}
}

func TestCreateMenuCommand_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cmd     CreateMenuCommand
		wantErr error
	}{
		{
			name: "Creation Root menu failed due to missing Name",
			cmd: CreateMenuCommand{
				ParentID: 11, // is sub, Assume the ID of the last added menu is 11
				Code:     "deleted user page",
				//Name:     "User Deleted Management",
				Path: "/user/deleted", // status=deleted the user list
				Icon: "user",
				Sort: 10,
			},
			wantErr: errors.New("code and name are required"),
		}, {
			name: "Creation Root menu failed due to missing Code",
			cmd: CreateMenuCommand{
				ParentID: 11, // is sub, Assume the ID of the last added menu is 11
				//Code:     "deleted user page",
				Name: "User Deleted Management",
				Path: "/user/deleted", // status=deleted the user list
				Icon: "user",
				Sort: 10,
			},
			wantErr: errors.New("code and name are required"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Validate()
			require.Error(t, err)
			assert.EqualError(t, err, tt.wantErr.Error())
		})
	}
}

// goos: darwin
// goarch: arm64
// pkg: github.com/Crows-Storm/Axis/auth/app/command
// cpu: Apple M5
// BenchmarkCreateMenuCommandHandler
// BenchmarkCreateMenuCommandHandler-10    	   25258	     45615 ns/op
// PASS
func BenchmarkCreateMenuCommandHandler(b *testing.B) {
	repo := new(menumock.MockRepository)
	mc := new(decoratormock.MockMetricsClient)
	bus := new(decoratormock.MockEventBus)

	// mock Expectations
	repo.On("Create", mock.Anything, mock.AnythingOfType("*menu.Menu")).Return(nil) // error is empty
	mc.On("Inc", mock.Anything, mock.Anything)
	bus.On("Publish", context.Background(), mock.Anything).Return(nil)

	cmd := CreateMenuCommand{
		ParentID: 0,
		Code:     "user",
		Name:     "User Management",
		Path:     "/user",
		Icon:     "user",
		Sort:     10,
	}

	handler := NewCreateMenuCommandHandler(repo, mc, bus)

	// Act
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.Handle(context.Background(), cmd)
	}
}

func TestUpdateMenuCommandHandler(t *testing.T) {

}
