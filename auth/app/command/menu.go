package command

import (
	"context"

	"github.com/Crows-Storm/Axis/auth/domain/menu"
	"github.com/Crows-Storm/Axis/common/decorator"
)

type CreateMenuCommand struct {
	ParentID uint64 `json:"parent_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
}

func (c CreateMenuCommand) Validate() decorator.CommandError {
	if c.Code == "" || c.Name == "" {
		return decorator.InvalidCommand("code and name are required")
	}
	return nil
}

type CreateMenuCommandHandler decorator.CommandHandler[CreateMenuCommand, *menu.Menu]

func NewCreateMenuCommandHandler(menuRepo menu.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) CreateMenuCommandHandler {
	if menuRepo == nil {
		panic("nil Menu Repository")
	}
	return decorator.ApplyCommandDecorators[CreateMenuCommand, *menu.Menu](
		createMenuCommandHandler{menuRepo: menuRepo, bus: bus},
		metricsClient,
	)
}

type createMenuCommandHandler struct {
	menuRepo menu.Repository
	bus      decorator.EventBus
}

func (h createMenuCommandHandler) Handle(ctx context.Context, cmd CreateMenuCommand) (*menu.Menu, error) {
	created, err := menu.New(cmd.ParentID, cmd.Code, cmd.Name, cmd.Path, cmd.Icon, cmd.Sort)
	if err != nil {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.menuRepo.Create(ctx, created); err != nil {
		return nil, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishMenu(ctx, h.bus, "MenuCreated", created); err != nil {
		return nil, err
	}
	return created, nil
}

type UpdateMenuCommand struct {
	MenuID   uint64 `json:"menu_id"`
	ParentID uint64 `json:"parent_id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	Sort     int    `json:"sort"`
}

func (c UpdateMenuCommand) Validate() decorator.CommandError {
	if c.MenuID == 0 || c.Name == "" {
		return decorator.InvalidCommand("menu_id and name are required")
	}
	return nil
}

type UpdateMenuCommandHandler decorator.CommandHandler[UpdateMenuCommand, struct{}]

func NewUpdateMenuCommandHandler(menuRepo menu.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) UpdateMenuCommandHandler {
	if menuRepo == nil {
		panic("nil Menu Repository")
	}
	return decorator.ApplyCommandDecorators[UpdateMenuCommand, struct{}](
		updateMenuCommandHandler{menuRepo: menuRepo, bus: bus},
		metricsClient,
	)
}

type updateMenuCommandHandler struct {
	menuRepo menu.Repository
	bus      decorator.EventBus
}

func (h updateMenuCommandHandler) Handle(ctx context.Context, cmd UpdateMenuCommand) (struct{}, error) {
	current, err := h.menuRepo.Get(ctx, cmd.MenuID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.Update(cmd.ParentID, cmd.Name, cmd.Path, cmd.Icon, cmd.Sort); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.menuRepo.Update(ctx, current); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishMenu(ctx, h.bus, "MenuUpdated", current); err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}

type UpdateMenuStatusCommand struct {
	MenuID uint64 `json:"menu_id"`
	Status int8   `json:"status"`
}

func (c UpdateMenuStatusCommand) Validate() decorator.CommandError {
	if c.MenuID == 0 || (c.Status != 0 && c.Status != 1) {
		return decorator.InvalidCommand("menu_id and status (0|1) are required")
	}
	return nil
}

type UpdateMenuStatusCommandHandler decorator.CommandHandler[UpdateMenuStatusCommand, struct{}]

func NewUpdateMenuStatusCommandHandler(menuRepo menu.Repository, metricsClient decorator.MetricsClient, bus decorator.EventBus) UpdateMenuStatusCommandHandler {
	if menuRepo == nil {
		panic("nil Menu Repository")
	}
	return decorator.ApplyCommandDecorators[UpdateMenuStatusCommand, struct{}](
		updateMenuStatusCommandHandler{menuRepo: menuRepo, bus: bus},
		metricsClient,
	)
}

type updateMenuStatusCommandHandler struct {
	menuRepo menu.Repository
	bus      decorator.EventBus
}

func (h updateMenuStatusCommandHandler) Handle(ctx context.Context, cmd UpdateMenuStatusCommand) (struct{}, error) {
	current, err := h.menuRepo.Get(ctx, cmd.MenuID)
	if err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := current.ChangeStatus(cmd.Status); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := h.menuRepo.Update(ctx, current); err != nil {
		return struct{}{}, decorator.CommandExecutedError{Msg: err.Error()}
	}
	if err := publishMenu(ctx, h.bus, "MenuStatusChanged", current); err != nil {
		return struct{}{}, err
	}
	return struct{}{}, nil
}
