package command

import (
	"context"
	"testing"

	"github.com/Crows-Storm/Axis/auth/domain/permission"
	"github.com/Crows-Storm/Axis/auth/domain/role"
	rolemock "github.com/Crows-Storm/Axis/auth/domain/role/mocks"
	"github.com/Crows-Storm/Axis/common/domain/event"
	commrole "github.com/Crows-Storm/Axis/common/domain/role"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type nopMetrics struct{}

func (nopMetrics) Inc(string, int) {}

func TestUnassignUserRoleCommand(t *testing.T) {
	repo := rolemock.NewMockRepository(t)
	current := role.Reconstruct(1, "editor", "编辑", "", commrole.Active, nil)
	repo.EXPECT().Get(mock.Anything, uint64(1)).Return(current, nil)
	repo.EXPECT().UnassignUser(mock.Anything, uint64(2), mock.AnythingOfType("*role.Role")).Return(nil)

	h := NewUnassignUserRoleCommandHandler(repo, nopMetrics{}, nil)
	_, err := h.Handle(context.Background(), UnassignUserRoleCommand{UserID: 2, RoleID: 1})
	require.NoError(t, err)
}

func TestCreateRoleCommandRejectsEmptyName(t *testing.T) {
	require.NotNil(t, CreateRoleCommand{Code: "x"}.Validate())
}

func TestRevokeRoleMenuCommand(t *testing.T) {
	repo := rolemock.NewMockRepository(t)
	current := role.Reconstruct(1, "editor", "编辑", "", commrole.Active, map[uint64]permission.Mask{
		9: permission.PermRead,
	})
	repo.EXPECT().Get(mock.Anything, uint64(1)).Return(current, nil)
	repo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*role.Role"), mock.Anything).Return(nil)

	h := NewRevokeRoleMenuCommandHandler(repo, nopMetrics{}, nil)
	_, err := h.Handle(context.Background(), RevokeRoleMenuCommand{RoleID: 1, MenuID: 9})
	require.NoError(t, err)
}

func TestCreateRoleCommand(t *testing.T) {
	repo := rolemock.NewMockRepository(t)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*role.Role")).Return(nil)

	h := NewCreateRoleCommandHandler(repo, nopMetrics{}, nil)
	created, err := h.Handle(context.Background(), CreateRoleCommand{Code: "ops", Name: "运维"})
	require.NoError(t, err)
	require.Equal(t, "ops", created.Code)
}

type captureBus struct {
	names []string
}

func (b *captureBus) Publish(_ context.Context, events ...event.Envelope[event.Message]) error {
	for _, evt := range events {
		if evt.Message != nil {
			b.names = append(b.names, evt.Message.Name())
		}
	}
	return nil
}

func (b *captureBus) Close() error { return nil }

func TestCreateRoleCommandPublishesRoleCreated(t *testing.T) {
	repo := rolemock.NewMockRepository(t)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*role.Role")).Return(nil)
	bus := &captureBus{}

	h := NewCreateRoleCommandHandler(repo, nopMetrics{}, bus)
	_, err := h.Handle(context.Background(), CreateRoleCommand{Code: "ops", Name: "运维"})
	require.NoError(t, err)
	require.Equal(t, []string{"RoleCreated"}, bus.names)
}
