package controller

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func ToolsTree(c *gin.Context) {
	DB := common.GetDB()

	var tools model.ToolSlice
	DB.Order("orders asc").Find(&tools)
	tools = tools.BuildTree(0)

	response.Success(c, gin.H{"tools": tools.ConvertDto()}, "ok")
}

func ToolsSearch(c *gin.Context) {
	search := c.PostForm("search")

	// 基本预处理：删除多余空格，限制长度
	search = strings.TrimSpace(search)
	if len(search) > 100 {
		response.Fail(c, nil, "搜索词过长")
		return
	}

	var tools model.ToolSlice

	if len(search) == 0 {
		response.Success(c, gin.H{"tools": tools}, "ok")
		return
	}

	DB := common.GetDB()
	DB.Where("title LIKE ? AND url != '' ", "%"+search+"%").Order("orders asc").Find(&tools)

	response.Success(c, gin.H{"tools": tools.ConvertDto()}, "ok")
}
