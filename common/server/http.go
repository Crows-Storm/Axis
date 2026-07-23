package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RunHTTPServer(name string, wrapper func(router *gin.Engine)) {
	addr := viper.Sub(name).GetString("http-addr")

	RunHTTPServerOnAddr(addr, wrapper)

}

func RunHTTPServerOnAddr(addr string, wrapper func(router *gin.Engine)) {
	apiRouter := gin.New()
	apiRouter.GET("/api/ping", func(c *gin.Context) {
		c.JSON(200, "pone")
	})
	wrapper(apiRouter)

	if err := apiRouter.Run(addr); err != nil {
		panic(err)
	}
}
