package routes

import (
	"jugg-tool-box-service/controller"
	"jugg-tool-box-service/middleware"

	"github.com/gin-gonic/gin"
)

func CollectRoute(r *gin.Engine) *gin.Engine {
	r.POST("/api/auth/regist", controller.Regist)
	r.POST("/api/auth/sign-in", controller.SignIn)
	r.POST("/api/auth/set-captcha", controller.SetCaptcha)
	r.POST("/api/auth/validate-captcha", controller.ValidateCaptcha)
	r.POST("/api/auth/send-email-code", controller.SendEmailCode)
	r.POST("/api/auth/info", middleware.AuthMiddleware(), controller.Info)
	r.POST("/api/auth/change-password", middleware.AuthMiddleware(), controller.ChangePassword)

	r.POST("/api/tools/tree", controller.ToolsTree)
	r.POST("/api/tools/search", controller.ToolsSearch)

	r.POST("/api/img/resize", middleware.AuthMiddleware(), controller.Resize)
	r.POST("/api/img/txt2img", middleware.AuthMiddleware(), controller.Txt2img)
	r.POST("/api/img/rembg", middleware.AuthMiddleware(), controller.Rembg)
	r.POST("/api/img/img2img", middleware.AuthMiddleware(), controller.Img2img)
	r.POST("/api/img/count-points", controller.CountPoints)

	r.POST("/api/tasks/task", controller.Task)
	r.POST("/api/tasks", middleware.AuthMiddleware(), controller.Tasks)
	r.POST("/api/tasks/home", controller.TasksInHome)
	r.POST("/api/points", middleware.AuthMiddleware(), controller.Points)

	r.POST("/api/seo", controller.Seo)

	// web socket 路由
	r.GET("/sd-callback", controller.SDCallbackHandler)
	r.GET("/progress", controller.ClientProgressHandler)

	// 将 /images 文件夹暴露为静态文件
	r.GET("/images/*filepath", controller.StaticImage)
	return r
}
