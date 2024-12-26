package controller

import (
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/util"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Points(c *gin.Context) {
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

	var points model.PointSlice
	db := common.GetDB()

	// 获取不分页的记录数
	var total int64
	db.Model(&model.Point{}).Where("uid = ? ", user.UID).Count(&total)

	// 查询分页的数据
	db.Where("uid = ? ", user.UID).Order("id desc").Offset(form.Offset).Limit(form.Limit).Find(&points)

	// todo 获取不进行分页的记录数
	response.Success(c, gin.H{"points": points.ConvertDto(), "total": total}, "ok")
}

func CountPoints(c *gin.Context) {
	widthStr := c.PostForm("width")
	heightStr := c.PostForm("height")
	numStr := c.PostForm("num")

	width, err := strconv.Atoi(widthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid width"})
		return
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid height"})
		return
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid num"})
		return
	}

	response.Success(c, gin.H{"points": countPoints(width, height, num)}, "ok")
}

// 计算需要消耗的点数
func countPoints(w int, h int, num int) int {
	floatCount := (float64(w) + float64(h)) / 1000
	count := int(math.Round(floatCount))

	count = count * num

	if count <= 0 {
		count = 1
	}

	return int(count)
}
