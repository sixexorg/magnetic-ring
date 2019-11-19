package router

import (
	"github.com/gin-gonic/gin"
)

var (
 	router *gin.Engine
)

func GetRouter() *gin.Engine {
	if router == nil {
		router = gin.Default()
	}

	return router
}
