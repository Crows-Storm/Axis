package server

import (
	"github.com/gin-gonic/gin"
)

func RunHTTPServer(addr string, wrapper func(router *gin.Engine)) {
	RunHTTPServerOnAddr(addr, wrapper)

}

func RunHTTPServerOnAddr(addr string, wrapper func(router *gin.Engine)) {
	apiRouter := gin.New()
	apiRouter.Any("/api/v1/ping", func(c *gin.Context) {
		Success(c, "pong")
	})
	wrapper(apiRouter)
	//apiRouter.Use(gin.Recovery())

	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}
