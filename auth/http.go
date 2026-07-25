package main

import (
	"github.com/Crows-Storm/Axis/auth/app"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	app app.Application
}

func (h HttpServer) AuthRoot(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h HttpServer) Login(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h HttpServer) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
