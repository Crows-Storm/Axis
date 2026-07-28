package main

import (
	"fmt"

	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/Crows-Storm/Axis/auth/app/command"
	"github.com/Crows-Storm/Axis/common/server"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) AuthRoot(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) Login(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (H HTTPServer) Register(c *gin.Context) {
	var req command.RegisterUserCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Invalid request body: %v", err))
		return
	}
	_, err := H.app.Commands.RegisterUser.Handle(c, req)
	if err != nil {
		server.Error(c, server.CodeServerError, fmt.Sprintf("Faield: %v", err))
		return
	}
	server.Success(c, "successfully")
	return
}
