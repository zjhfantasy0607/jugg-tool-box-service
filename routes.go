package main

import (
	"jugg-tool-box-service/controller"
	"jugg-tool-box-service/middleware"

	"github.com/gin-gonic/gin"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	r.POST("/api/auth/regist", controller.Regist)
	r.POST("/api/auth/signIn", controller.SignIn)
	r.POST("/api/auth/setCaptcha", controller.SetCaptcha)
	r.POST("/api/auth/validateCaptcha", controller.ValidateCaptcha)

	r.POST("/api/auth/info", middleware.AuthMiddleware(), controller.Info)

	return r
}
