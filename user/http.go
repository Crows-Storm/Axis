package main

import (
	"context"
	time "time"

	"github.com/Crows-Storm/Axis/common/config/logger"
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
}

func (H HTTPServer) Disable(c *gin.Context, id int64) {
	result, err := H.app.Commands.DisableUser.Handle(c, command.DisableUserCommand{
		Id: id,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	server.Success(c, result)
}

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
	var req query.UserExists
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
	}
	result, err := H.app.Queries.UserExists.Handle(c, req)
	if err != nil {
		logger.Infof("Failed to get user stats: %v", err)
	}
	server.Success(c, result)
}

// UserStatusAnalysis Query the number of users in all states
func (H HTTPServer) UserStatusAnalysis(c *gin.Context) {
	//val, ok := c.Get("session_holder_id")
	//if !ok {
	//	server.ErrorWithCode(c, server.CodeBadRequest, "session_holder_id not found in context")
	//	return
	//}
	//sessionHolderID, _ := val.(int64)
	result, err := H.app.Queries.UserStatusAnalysis.Handle(c, query.GetUserStatsQuery{})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) SoftDeleteUser(c *gin.Context, id int64) {
	_, err := H.app.Commands.SoftDeleteUser.Handle(c, command.SoftDeleteUserCommand{
		Id: id,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, true)
}

func (H HTTPServer) UpdateCurrentUserInfo(c *gin.Context) {
	val, ok := c.Get("session_holder_id")
	if !ok {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	sessionHolderID, _ := val.(int64)

	// bind to domain object
	var req domain.User
	if err := c.ShouldBindJSON(&req); err != nil {
		server.ErrorWithCode(c, server.CodeBadRequest)
		return
	}
	// setting holder Id to updated
	req.Id = sessionHolderID
	// just have email and id in req now
	_, err := H.app.Commands.UpdateUser.Handle(c, command.UpdateUserCommand{
		User: &req, // from db by request context get user id to query a user domain
		UpdateFun: func(ctx context.Context, u *domain.User) (*domain.User, error) {
			u.UpdateTime = time.Now()
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
	// TODO: need add a Context Manager for gin context
	// is current session holder id
	val, ok := c.Get("session_holder_id")
	if !ok {
		server.ErrorWithStatusUnauthorized(c)
		return
	}
	sessionHolderID, _ := val.(int64)
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		Id: sessionHolderID,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}

func (H HTTPServer) GetUserInfoById(c *gin.Context, id int64) {
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		Id: id,
	})
	if err != nil {
		server.ErrorWithCode(c, server.CodeInternalServerError)
		return
	}
	server.Success(c, result)
}
