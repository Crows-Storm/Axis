package main

import (
	"context"

	"github.com/Crows-Storm/Axis/common/config/logger"
	"github.com/Crows-Storm/Axis/common/contextx"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	domain "github.com/Crows-Storm/Axis/user/domain/user"
	"github.com/Crows-Storm/Axis/user/protos"
	"github.com/gin-gonic/gin"
)

// HTTPServer all command not return data, but query can return something
type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) CreateUser(c *gin.Context) {
	var req command.CreateUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}

	result, err := H.app.Commands.CreateUser.Handle(c, req)
	if err != nil {
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, result)
	return
}

func (H HTTPServer) Disable(c *gin.Context, id uint64) {
	result, err := H.app.Commands.DisableUser.Handle(c, command.DisableUserCommand{
		ID: id,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	server.Success(c, result)
}

// The CreateBatchUsers Deprecated
func (H HTTPServer) CreateBatchUsers(c *gin.Context) {
	var req command.CreateBatchUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}

	result, err := H.app.Commands.CreateBatchUsers.Handle(c, req)
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) ExistsWithTransaction(c *gin.Context, params protos.ExistsWithTransactionParams) {
	req := query.UserExists{}
	if params.Id != nil {
		req.ID = uint64(*params.Id)
	}
	if params.LoginId != nil {
		req.LoginID = *params.LoginId
	}
	if params.Email != nil {
		req.Email = *params.Email
	}
	result, err := H.app.Queries.UserExists.Handle(c, req)
	if err != nil {
		logger.Infof("Failed to check user exists: %v", err)
		server.ErrorWithCodeAndMessage(c, server.CodeBadRequest, err.Error())
		return
	}
	server.Success(c, result)
}

// UserStatusAnalysis Query the number of users in all states
func (H HTTPServer) UserStatusAnalysis(c *gin.Context) {
	result, err := H.app.Queries.UserStatusAnalysis.Handle(c, query.GetUserStatsQuery{})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) SoftDeleteUser(c *gin.Context, id uint64) {
	_, err := H.app.Commands.SoftDeleteUser.Handle(c, command.SoftDeleteUserCommand{
		ID: id,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, true)
}

func (H HTTPServer) UpdateCurrentUserInfo(c *gin.Context) {
	reader := contextx.FromGinContext(c)
	if !reader.IsAuthenticated() {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}

	// bind to domain object
	var req domain.User
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	req.ID = reader.UserID()
	// just have email and id in req now
	_, err := H.app.Commands.UpdateUser.Handle(c, command.UpdateUserCommand{
		User: &req, // from db by request contextx get user id to query a user domain
		UpdateFun: func(ctx context.Context, u *domain.User) (*domain.User, error) {
			if &req != nil {
				if req.Email != "" {
					u.Email = req.Email
				}
			}
			return u, nil
		},
	})

	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, true)
	return
}

func (H HTTPServer) GetCurrentUserInfo(c *gin.Context) {
	reader := contextx.FromGinContext(c)
	if !reader.IsAuthenticated() {
		server.ErrorWithStatusUnauthorized(c)
		return
	}
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		ID: reader.UserID(),
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) GetUserInfoById(c *gin.Context, id uint64) {
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		ID: id,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}
