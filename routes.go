package main

import (
	"jugg-tool-box-service/controller"
	"jugg-tool-box-service/middleware"

	"github.com/gin-gonic/gin"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	r.POST("/api/auth/regist", controller.Regist)
	r.POST("/api/auth/signIn", controller.SignIn)
	r.GET("/api/auth/info", middleware.AuthMiddleware(), controller.Info)

	return r
}
