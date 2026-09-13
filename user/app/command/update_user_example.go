// Package command provides example implementations for user update commands
package command

//import (
//	"context"
//	"fmt"
//
//	"github.com/Crows-Storm/Axis/user/domain/user"
//)
//
//// UpdateUserCommand2 更新用户命令
//type UpdateUserCommand2 struct {
//	UserID  int64
//	Email   *string
//	LoginID *string
//	Status  *int8
//}
//
//// UpdateUserHandler 更新用户命令处理器
//type UpdateUserHandler struct {
//	userRepo user.Repository
//}
//
//// NewUpdateUserHandler 创建命令处理器
//func NewUpdateUserHandler(userRepo user.Repository) *UpdateUserHandler {
//	return &UpdateUserHandler{
//		userRepo: userRepo,
//	}
//}
//
//// Handle 处理更新命令
//func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand2) error {
//	// 1. 加载聚合根
//	u, err := h.userRepo.GetInfo(cmd.UserID)
//	if err != nil {
//		return fmt.Errorf("user not found: %w", err)
//	}
//
//	// 2. 构建更新选项
//	var options []user.UpdateOption
//	if cmd.Email != nil {
//		options = append(options, user.UpdateOptionWithEmail(*cmd.Email))
//	}
//	if cmd.LoginID != nil {
//		options = append(options, user.WithLoginID(*cmd.LoginID))
//	}
//	if cmd.Status != nil {
//		options = append(options, user.WithStatus(*cmd.Status))
//	}
//
//	// 3. 执行更新
//	if err := u.Update(ctx, options...); err != nil {
//		return fmt.Errorf("failed to update user: %w", err)
//	}
//
//	// 4. 保存（会自动发布事件）
//	// Note: 需要根据实际的Repository接口调整
//	// 当前的Repository接口使用了不同的签名
//	updateFunc := func(ctx context.Context, usr *user.User) (*user.User, error) {
//		return usr, nil
//	}
//
//	if err := h.userRepo.Update(ctx, u, updateFunc); err != nil {
//		return fmt.Errorf("failed to save user: %w", err)
//	}
//
//	return nil
//}
//
//// ChangeEmailCommand 修改邮箱命令
//type ChangeEmailCommand struct {
//	UserID   int64
//	NewEmail string
//}
//
//// ChangeEmailHandler 修改邮箱命令处理器
//type ChangeEmailHandler struct {
//	userRepo user.Repository
//}
//
//// NewChangeEmailHandler 创建命令处理器
//func NewChangeEmailHandler(userRepo user.Repository) *ChangeEmailHandler {
//	return &ChangeEmailHandler{
//		userRepo: userRepo,
//	}
//}
//
//// Handle 处理修改邮箱命令
//func (h *ChangeEmailHandler) Handle(ctx context.Context, cmd ChangeEmailCommand) error {
//	// 1. 加载用户
//	u, err := h.userRepo.GetInfo(cmd.UserID)
//	if err != nil {
//		return fmt.Errorf("user not found: %w", err)
//	}
//
//	// 2. 执行专用业务行为（包含业务规则验证）
//	if err := u.ChangeEmail(ctx, cmd.NewEmail); err != nil {
//		return fmt.Errorf("failed to change email: %w", err)
//	}
//
//	// 3. 保存（会发布 user.email_changed 事件）
//	updateFunc := func(ctx context.Context, usr *user.User) (*user.User, error) {
//		return usr, nil
//	}
//
//	if err := h.userRepo.Update(ctx, u, updateFunc); err != nil {
//		return fmt.Errorf("failed to save user: %w", err)
//	}
//
//	return nil
//}
//
//// BatchUpdateUsersCommand 批量更新用户命令
//type BatchUpdateUsersCommand struct {
//	Updates []UpdateUserCommand2
//}
//
//// BatchUpdateUsersHandler 批量更新命令处理器
//type BatchUpdateUsersHandler struct {
//	userRepo user.Repository
//}
//
//// NewBatchUpdateUsersHandler 创建批量更新处理器
//func NewBatchUpdateUsersHandler(userRepo user.Repository) *BatchUpdateUsersHandler {
//	return &BatchUpdateUsersHandler{
//		userRepo: userRepo,
//	}
//}
//
//// Handle 处理批量更新命令
//func (h *BatchUpdateUsersHandler) Handle(ctx context.Context, cmd BatchUpdateUsersCommand) error {
//	users := make([]*user.User, 0, len(cmd.Updates))
//
//	// 加载并更新所有用户
//	for _, updateCmd := range cmd.Updates {
//		u, err := h.userRepo.GetInfo(updateCmd.UserID)
//		if err != nil {
//			return fmt.Errorf("user %d not found: %w", updateCmd.UserID, err)
//		}
//
//		// 构建更新选项
//		var options []user.UpdateOption
//		if updateCmd.Email != nil {
//			options = append(options, user.UpdateOptionWithEmail(*updateCmd.Email))
//		}
//		if updateCmd.LoginID != nil {
//			options = append(options, user.WithLoginID(*updateCmd.LoginID))
//		}
//		if updateCmd.Status != nil {
//			options = append(options, user.WithStatus(*updateCmd.Status))
//		}
//
//		// 执行更新
//		if err := u.Update(ctx, options...); err != nil {
//			return fmt.Errorf("failed to update user %d: %w", updateCmd.UserID, err)
//		}
//
//		users = append(users, u)
//	}
//
//	// 批量保存
//	if err := h.userRepo.CreateBatch(ctx, users); err != nil {
//		return fmt.Errorf("failed to batch save users: %w", err)
//	}
//
//	return nil
//}
//
//// Example usage functions
//
//// ExampleUpdateSingleField 示例：更新单个字段
//func ExampleUpdateSingleField(ctx context.Context, repo user.Repository, userID int64) error {
//	u, err := repo.GetInfo(userID)
//	if err != nil {
//		return err
//	}
//
//	// 只更新邮箱
//	newEmail := "newemail@example.com"
//	return u.Update(ctx, user.UpdateOptionWithEmail(newEmail))
//}
//
//// ExampleUpdateMultipleFields 示例：更新多个字段
//func ExampleUpdateMultipleFields(ctx context.Context, repo user.Repository, userID int64) error {
//	u, err := repo.GetInfo(userID)
//	if err != nil {
//		return err
//	}
//
//	// 同时更新多个字段
//	newEmail := "newemail@example.com"
//	newLoginID := "newlogin"
//	newStatus := user.StatusActive
//
//	return u.Update(ctx,
//		user.UpdateOptionWithEmail(newEmail),
//		user.WithLoginID(newLoginID),
//		user.WithStatus(newStatus),
//	)
//}
//
//// ExampleUpdateWithSnapshot 示例：使用快照更新
//func ExampleUpdateWithSnapshot(ctx context.Context, repo user.Repository, userID int64) error {
//	u, err := repo.GetInfo(userID)
//	if err != nil {
//		return err
//	}
//
//	// 创建快照（例如从 DTO 映射而来）
//	snapshot := &user.User{
//		ID:      u.ID,
//		LoginId: "updatedlogin",
//		Email:   "updated@example.com",
//		Status:  user.StatusActive,
//	}
//
//	return u.UpdateWithSnapshot(ctx, snapshot)
//}
//
//// ExampleChangeEmailWithBusinessRules 示例：使用专用业务方法
//func ExampleChangeEmailWithBusinessRules(ctx context.Context, repo user.Repository, userID int64) error {
//	u, err := repo.GetInfo(userID)
//	if err != nil {
//		return err
//	}
//
//	// 使用专用方法，包含业务规则验证
//	// - 检查邮箱是否相同
//	// - 检查用户是否激活
//	// - 发布专用的 user.email_changed 事件
//	return u.ChangeEmail(ctx, "new@example.com")
//}
