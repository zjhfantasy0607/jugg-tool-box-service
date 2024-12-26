package controller

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/util"
	"strings"

	"github.com/gin-gonic/gin"
)

func Seo(c *gin.Context) {
	url := c.PostForm("url")

	// 基本预处理：删除多余空格，限制长度
	url = strings.TrimSpace(url)

	// 检查 url 是否为空
	if url == "" {
		response.Fail(c, nil, "url不能为空")
		return
	}

	// 获取数据库连接
	DB := common.GetDB()

	// 查询指定的 seo 记录，url 开头匹配数据库中的 url
	var seo model.Seo
	if err := DB.Where("url IN ?", util.SplitLevelPath(url)).Order("CHAR_LENGTH(url) DESC").First(&seo).Error; err != nil {
		// 如果没有找到记录或发生错误
		response.Fail(c, nil, "记录不存在")
		return
	}

	// 返回查询结果
	response.Success(c, gin.H{"seo": seo.ConvertDto()}, "ok")
}
