package server

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func RunHTTPServer(name string, wrapper func(router *gin.Engine)) {
	addr := viper.Sub(name).GetString("http-addr")

	RunHTTPServerOnAddr(addr, wrapper)

}

func RunHTTPServerOnAddr(name string, wrapper func(router *gin.Engine)) {

}
