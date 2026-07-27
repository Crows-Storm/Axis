package main

import (
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/command"
	"github.com/Crows-Storm/Axis/user/app/query"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
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
	// TODO: the JWT implementation should include retrieving the ID from the current user context
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		Id: 123,
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
