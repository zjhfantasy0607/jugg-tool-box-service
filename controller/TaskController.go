package controller

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/util"

	"github.com/gin-gonic/gin"
)

func Task(c *gin.Context) {
	taskId := c.PostForm("task_id")

	db := common.GetDB()
	var task model.Task
	if err := db.Where("task_id = ?", taskId).First(&task).Error; err != nil {
		response.Fail(c, gin.H{}, "Task not found")
		return
	}

	response.Success(c, gin.H{"task": task.ConvertDto()}, "ok")
}

func Tasks(c *gin.Context) {
	// 绑定并验证表单数据
	var form struct {
		Offset int `form:"offset"`
		Limit  int `form:"limit"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	userData, _ := c.Get("user")
	user := userData.(model.User)

	var tasks model.TaskSlice
	db := common.GetDB()

	// 获取不分页的记录数
	var total int64
	db.Model(&model.Task{}).Where("uid = ? ", user.UID).Count(&total)

	// 查询分页的数据
	db.Where("uid = ? ", user.UID).Order("id desc").Offset(form.Offset).Limit(form.Limit).Find(&tasks)

	// 查询 tool
	var tools model.ToolSlice
	db.Order("orders asc").Find(&tools)

	// 将 tools 转为 map[string]model.Tool
	toolMap := make(map[string]*model.Tool)
	for _, tool := range tools {
		toolMap[tool.Tool] = tool
	}

	// 添加中文的工具名
	tasksDto := tasks.ConvertDto()
	for i, task := range tasksDto {
		tasksDto[i].ToolTitle = toolMap[task.Tool].Title
		tasksDto[i].ToolUrl = toolMap[task.Tool].Url
	}

	// todo 获取不进行分页的记录数

	response.Success(c, gin.H{"tasks": tasksDto, "total": total}, "ok")
}

func TasksInHome(c *gin.Context) {
	// 绑定并验证表单数据
	var form struct {
		Offset int `form:"offset"`
		Limit  int `form:"limit"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	db := common.GetDB()
	var tasks model.TaskSlice

	// 获取不分页的记录数
	var total int64
	db.Model(&model.Task{}).Where("tool IN (?)", []string{"img2img", "txt2img"}).Where("status = ?", "success").Count(&total)

	// 查询分页的数据
	db.Where("tool IN (?)", []string{"img2img", "txt2img"}).Order("id desc").Where("status = ?", "success").Offset(form.Offset).Limit(form.Limit).Find(&tasks)

	// 查询 tool
	var tools model.ToolSlice
	db.Order("orders asc").Find(&tools)

	// 将 tools 转为 map[string]model.Tool
	toolMap := make(map[string]*model.Tool)
	for _, tool := range tools {
		toolMap[tool.Tool] = tool
	}

	// 添加中文的工具名
	tasksDto := tasks.ConvertDto()
	for i, task := range tasksDto {
		tasksDto[i].ToolTitle = toolMap[task.Tool].Title
		tasksDto[i].ToolUrl = toolMap[task.Tool].Url
	}

	// todo 获取不进行分页的记录数
	response.Success(c, gin.H{"tasks": tasksDto, "total": total}, "ok")
}
