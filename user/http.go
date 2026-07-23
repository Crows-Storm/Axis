package main

import (
	"net/http"

	"github.com/Crows-Storm/Axis/user/app"
	"github.com/Crows-Storm/Axis/user/app/query"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H HTTPServer) GetCurrentUserInfo(c *gin.Context) {
	result, err := H.app.Queries.GetUser.Handle(c, query.GetUserQuery{
		Id: 123,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Failed", "error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": result})
}
