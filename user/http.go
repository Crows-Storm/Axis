package main

import (
	"fmt"

	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) CreateBatchUsers(c *gin.Context) {
	var req []command.CreateBatchUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Error(c, server.CodeBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	result, err := H.app.Commands.CreateBatchUsers.Handle(c, req)
	if err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Failed to create batch users: %v", err))
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) ExistsWithTransaction(c *gin.Context, params protos.ExistsWithTransactionParams) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) ListUsers(c *gin.Context, params protos.ListUsersParams) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) GetUserStats(c *gin.Context) {
	result, err := H.app.Queries.GetUserStats.Handle(c, query.GetUserStatsQuery{})
	if err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Failed to get user stats: %v", err))
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) SoftDeleteUser(c *gin.Context, id int64) {
	_, err := H.app.Commands.SoftDeleteUser.Handle(c, command.SoftDeleteUserCommand{
		ID: id,
	})
	if err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Failed to soft delete user: %v", err))
		return
	}
	server.Success(c, true)
}

func (H HTTPServer) UpdateUserStatus(c *gin.Context, id int64) {
	var req struct {
		Status int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Error(c, server.CodeBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	_, err := H.app.Commands.UpdateUserStatus.Handle(c, command.UpdateUserStatusCommand{
		ID:     id,
		Status: req.Status,
	})
	if err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Failed to update user status: %v", err))
		return
	}
	server.Success(c, true)
}

func (H HTTPServer) UpdateCurrentUserInfo(c *gin.Context) {
	val, ok := c.Get("session_holder_id")
	if !ok {
		server.Error(c, server.CodeBadRequest, "session_holder_id not found in context")
		return
	}
	sessionHolderID, _ := val.(int64)

	var req command.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Error(c, server.CodeBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	_, err := H.app.Commands.UpdateUser.Handle(c, command.UpdateUserCommand{
		UserID: sessionHolderID,
		UpdateFun: func(u *domain.User) error {
			if req.Username != "" {
				u.Username = req.Username
			}
			if req.Email != "" {
				u.Email = req.Email
			}
			return u.Validate()
		},
	})
	if err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Failed to update current user: %v", err))
		return
	}
	server.Success(c, true)
}

func (H HTTPServer) UpdateCurrentUserInfo(c *gin.Context) {
	_, err := H.app.Commands.UpdateUser.Handle(c, command.UpdateUserCommand{
		User:      nil, // from db by request context get user id to query a user domain
		UpdateFun: nil,
	})
	if err != nil {
		server.Error(c, server.CodeServerError, "Failed")
		return
	}
	server.Success(c, true)
	return
}

func (H HTTPServer) GetCurrentUserInfo(c *gin.Context) {
	// TODO: need add a Context Manager for gin context
	// is current session holder id
	val, ok := c.Get("session_holder_id")
	if !ok {
		c.JSON(400, gin.H{"error": "session_holder_id not found"})
		return
	}
	sessionHolderID, _ := val.(int64)
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		Id: sessionHolderID,
	})
	if err != nil {
		server.Error(c, server.CodeServerError, "Failed")
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) GetUserInfoById(c *gin.Context, id int64) {
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		Id: id,
	})
	if err != nil {
		server.Error(c, server.CodeServerError, "Failed")
		return
	}
	server.Success(c, result)
}
