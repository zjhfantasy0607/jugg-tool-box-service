package controller

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/sdconfig"
	"jugg-tool-box-service/util"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

type TopicMessage struct {
	TaskId string `json:"task_id"`
	Api    string `json:"api"`
	Params string `json:"params"`
}

func Resize(c *gin.Context) {
	var err error

	// 验证身份信息
	userInfo, _ := c.Get("user")
	user, ok := userInfo.(model.User)
	if !ok {
		response.Fail(c, nil, "身份校验失败")
		return
	}

	// 获取用户任务集合的长度
	length, err := common.Rdb.SCard(common.Ctx, user.UID).Result()
	if err != nil {
		errors.WithStack(err)
		response.Fail(c, nil, "查询任务错误")
		return
	}

	if length > viper.GetInt64("image.maxUserTask") {
		response.Fail(c, nil, "最多同时存在"+viper.GetString("image.maxUserTask")+"个任务")
		return
	}

	// 绑定并验证表单数据
	var form struct {
		InitImage string  `form:"init_image" binding:"required" label:"图片"`
		Resize    float64 `form:"resize" binding:"required,gte=1,max=3" label:"缩放大小"` // 必填，且大于等于1
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	form.InitImage = strings.TrimSpace(form.InitImage) // 过滤空字符

	// 保存图片到服务器并得到图片的宽高
	uploadPath := viper.GetString("image.upload")
	sourcePath, width, height, err := saveUploadImage(form.InitImage, uploadPath)
	if err != nil {
		response.Fail(c, nil, "上传图片存在错误")
		return
	}

	// 生成图的宽高设置为放大后的数值，用于计算需要消耗的点数
	width = int(float64(width) * form.Resize)
	height = int(float64(height) * form.Resize)
	if width > 5000 || height > 5000 {
		response.Fail(c, nil, "图片分辨率过高")
		return
	}

	// 获取 stable diffusion 的默认配置参数
	configJson := sdconfig.Get("resize")

	// 根据用户的设置参数对默认配置的参数进行修改
	configJson, err = util.SetJsonFields(configJson, map[string]any{
		"upscaling_resize": form.Resize,
		"image":            form.InitImage,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 创建任务 uuid
	taskId := uuid.New().String()

	// 创建要推入 Nsq 队列的Json消息
	msgJson, err := util.SetJsonFields("{}", map[string]any{
		"task_id": taskId,
		"api":     "sdapi/v1/extra-single-image",
		"params":  configJson,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户配置参数 json
	paramsJson, err := util.SetJsonFields("{}", map[string]any{"resize": form.Resize})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户上传图片文件路径 json
	sourceJson, err := util.SetJsonFields("[]", map[string]any{"0": sourcePath})
	if err != nil {
		response.Fail(c, nil, "参数错误")
		return
	}

	db := common.GetDB()

	// 计算要消耗的积分
	points := countPoints(width, height, 1)

	// 消耗积分
	pointRecord := model.Point{
		TaskId:       taskId,
		UID:          user.UID,
		BeforePoints: user.Points,
		Points:       points * -1,
		Tool:         "resize",
		Remark:       "图片高清修复",
	}
	err = pointRecord.AddRecord(db, &user)
	if err != nil {
		if err.Error() == "don't have enough points" {
			response.Fail(c, nil, "积分不足")
		} else {
			response.Response(c, http.StatusInternalServerError, 500, nil, "积分使用失败")
		}
		return
	}

	// 创建任务并推送到消息队列
	task := &model.Task{
		TaskId:     taskId,
		UID:        user.UID,
		Tool:       "resize",
		Params:     string(paramsJson),
		Source:     string(sourceJson),
		Output:     "[]",
		UsedPoints: points, // 根据生成图的宽高计算需要消耗的点数
		Status:     "pending",
		StartTime:  time.Now(),
	}

	err = createTask(task, &user, msgJson)
	if err != nil {
		util.LogErr(err, "./log/imageQueue.log")
		response.Response(c, http.StatusInternalServerError, 500, nil, "任务创建失败")
		return
	}

	response.Success(c, gin.H{"taskID": taskId}, "ok")
}

func Img2img(c *gin.Context) {
	var err error

	// 验证身份信息
	userInfo, _ := c.Get("user")
	user, ok := userInfo.(model.User)
	if !ok {
		response.Fail(c, nil, "身份校验失败")
		return
	}

	// 获取用户任务集合的长度
	length, err := common.Rdb.SCard(common.Ctx, user.UID).Result()
	if err != nil {
		errors.WithStack(err)
		response.Fail(c, nil, "查询任务错误")
		return
	}

	if length > viper.GetInt64("image.maxUserTask") {
		response.Fail(c, nil, "最多同时存在"+viper.GetString("image.maxUserTask")+"个任务")
		return
	}

	// 绑定并验证表单数据
	var form struct {
		InputImage string `form:"input_image" binding:"required" label:"图片"`
		Prompt     string `form:"prompt"  label:"提示词"`
		Width      int    `form:"width" binding:"required,numeric,gte=1" label:"宽度"`
		Height     int    `form:"height" binding:"required,numeric,gte=1" label:"高度"`
		Seed       int    `form:"seed" label:"随机数种子"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	form.InputImage = strings.TrimSpace(form.InputImage) // 过滤空字符

	// 保存图片到服务器并得到图片的宽高
	uploadPath := viper.GetString("image.upload")
	sourcePath, width, height, err := saveUploadImage(form.InputImage, uploadPath)
	if err != nil {
		response.Fail(c, nil, "上传图片存在错误")
		return
	}

	if width > 5000 || height > 5000 {
		response.Fail(c, nil, "图片分辨率过高")
		return
	}

	// 获取 stable diffusion 的默认配置参数
	configJson := sdconfig.Get("img2img")

	// 根据用户的设置参数对默认配置的参数进行修改
	configJson, err = util.SetJsonFields(configJson, map[string]any{
		"init_images": [1]string{form.InputImage},
		"prompt":      form.Prompt,
		"width":       form.Width,
		"height":      form.Height,
		"seed":        form.Seed,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 创建任务 uuid
	taskId := uuid.New().String()

	// 创建要推入 Nsq 队列的Json消息
	msgJson, err := util.SetJsonFields("{}", map[string]any{
		"task_id": taskId,
		"api":     "sdapi/v1/img2img",
		"params":  configJson,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户配置参数 json
	paramsJson, err := util.SetJsonFields("{}", map[string]any{
		"prompt": form.Prompt,
		"width":  form.Width,
		"height": form.Height,
		"seed":   form.Seed,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户上传图片文件路径 json
	sourceJson, err := util.SetJsonFields("[]", map[string]any{"0": sourcePath})
	if err != nil {
		response.Fail(c, nil, "参数错误")
		return
	}

	db := common.GetDB()

	// 计算要消耗的积分
	points := countPoints(width, height, 1)

	// 消耗积分
	pointRecord := model.Point{
		TaskId:       taskId,
		UID:          user.UID,
		BeforePoints: user.Points,
		Points:       points * -1,
		Tool:         "img2img",
		Remark:       "图生图",
	}
	err = pointRecord.AddRecord(db, &user)
	if err != nil {
		if err.Error() == "don't have enough points" {
			response.Fail(c, nil, "积分不足")
		} else {
			response.Response(c, http.StatusInternalServerError, 500, nil, "积分使用失败")
		}
		return
	}

	// 创建任务并推送到消息队列
	task := &model.Task{
		TaskId:     taskId,
		UID:        user.UID,
		Tool:       "img2img",
		Params:     string(paramsJson),
		Source:     string(sourceJson),
		Output:     "[]",
		UsedPoints: points, // 根据生成图的宽高计算需要消耗的点数
		Status:     "pending",
		StartTime:  time.Now(),
	}

	err = createTask(task, &user, msgJson)
	if err != nil {
		util.LogErr(err, "./log/imageQueue.log")
		response.Response(c, http.StatusInternalServerError, 500, nil, "任务创建失败")
		return
	}

	response.Success(c, gin.H{"taskID": taskId}, "ok")
}

func Txt2img(c *gin.Context) {
	var err error

	// 获取身份信息
	userInfo, _ := c.Get("user")
	user, ok := userInfo.(model.User)
	if !ok {
		response.Fail(c, nil, "身份校验失败")
		return
	}

	// 获取用户任务集合的长度
	length, err := common.Rdb.SCard(common.Ctx, user.UID).Result()
	if err != nil {
		errors.WithStack(err)
		response.Fail(c, nil, "查询任务错误")
		return
	}

	if length > viper.GetInt64("image.maxUserTask") {
		response.Fail(c, nil, "最多同时存在"+viper.GetString("image.maxUserTask")+"个任务")
		return
	}

	// 绑定并验证表单数据
	var form struct {
		Prompt string `form:"prompt"  label:"提示词"`
		Width  int    `form:"width" binding:"required,numeric,gte=1" label:"宽度"`
		Height int    `form:"height" binding:"required,numeric,gte=1" label:"高度"`
		Seed   int    `form:"seed" label:"随机数种子"`
	}
	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 获取 stable diffusion resize 的默认配置参数
	configJson := sdconfig.Get("txt2img")

	// 根据用户的设置参数对默认配置的参数进行修改
	configJson, err = util.SetJsonFields(configJson, map[string]any{
		"prompt": form.Prompt,
		"width":  form.Width,
		"height": form.Height,
		"seed":   form.Seed,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 创建任务 uuid
	taskId := uuid.New().String()

	// 创建要推入 Nsq 队列的Json消息
	msgJson, err := util.SetJsonFields("{}", map[string]any{
		"task_id": taskId,
		"api":     "sdapi/v1/txt2img",
		"params":  configJson,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户配置参数 json
	paramsJson, err := util.SetJsonFields("{}", map[string]any{
		"prompt": form.Prompt,
		"width":  form.Width,
		"height": form.Height,
		"seed":   form.Seed,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	db := common.GetDB()

	// 计算积分
	pointsMagnification := int(float32(len(form.Prompt))*0.008) + 1
	points := countPoints(form.Width, form.Height, pointsMagnification)

	// 消耗积分
	pointRecord := model.Point{
		TaskId:       taskId,
		UID:          user.UID,
		BeforePoints: user.Points,
		Points:       points * -1,
		Tool:         "resize",
		Remark:       "文生图",
	}
	err = pointRecord.AddRecord(db, &user)
	if err != nil {
		if err.Error() == "don't have enough points" {
			response.Fail(c, nil, "积分不足")
		} else {
			response.Response(c, http.StatusInternalServerError, 500, nil, "积分使用失败")
		}
		return
	}

	// 创建任务并推送到消息队列
	task := &model.Task{
		TaskId:     taskId,
		UID:        user.UID,
		Tool:       "txt2img",
		Params:     string(paramsJson),
		Source:     "[]",
		Output:     "[]",
		UsedPoints: points, // 根据生成图的宽高计算需要消耗的点数
		Status:     "pending",
		StartTime:  time.Now(),
	}

	err = createTask(task, &user, msgJson)
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Response(c, http.StatusInternalServerError, 500, nil, "任务创建失败")
		return
	}

	response.Success(c, gin.H{"taskID": taskId}, "ok")
}

func Rembg(c *gin.Context) {
	var err error

	// 验证身份信息
	userInfo, _ := c.Get("user")
	user, ok := userInfo.(model.User)
	if !ok {
		response.Fail(c, nil, "身份校验失败")
		return
	}

	// 获取用户任务集合的长度
	length, err := common.Rdb.SCard(common.Ctx, user.UID).Result()
	if err != nil {
		errors.WithStack(err)
		response.Fail(c, nil, "查询任务错误")
		return
	}

	if length > viper.GetInt64("image.maxUserTask") {
		response.Fail(c, nil, "最多同时存在"+viper.GetString("image.maxUserTask")+"个任务")
		return
	}

	// 绑定并验证表单数据
	var form struct {
		IsAnime    bool   `form:"is_anime" label:"是否动漫图片"`
		ReturnMask bool   `form:"return_mask" label:"是否返回蒙版"` // 必填，且大于等于1
		InputImage string `form:"input_image" binding:"required" label:"上传图片"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	form.InputImage = strings.TrimSpace(form.InputImage) // 过滤空字符

	// 保存图片到服务器并得到图片的宽高
	uploadPath := viper.GetString("image.upload")
	sourcePath, width, height, err := saveUploadImage(form.InputImage, uploadPath)
	if err != nil {
		response.Fail(c, nil, "上传图片存在错误")
		return
	}

	// 生成图的宽高设置为放大后的数值，用于计算需要消耗的点数
	if width > 5000 || height > 5000 {
		response.Fail(c, nil, "图片分辨率过高")
		return
	}

	// 获取 stable diffusion 的默认配置参数
	configJson := sdconfig.Get("rembg")

	// 根据用户的设置参数对默认配置的参数进行修改
	modelStr := "u2net"
	if form.IsAnime {
		modelStr = "isnet-anime"
	}

	configJson, err = util.SetJsonFields(configJson, map[string]any{
		"model":       modelStr,
		"return_mask": form.ReturnMask,
		"input_image": form.InputImage,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 创建任务 uuid
	taskId := uuid.New().String()

	// 创建要推入 Nsq 队列的Json消息
	msgJson, err := util.SetJsonFields("{}", map[string]any{
		"task_id": taskId,
		"api":     "rembg",
		"params":  configJson,
	})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户配置参数 json
	paramsJson, err := util.SetJsonFields("{}", map[string]any{"is_anime": form.IsAnime, "return_mask": form.ReturnMask})
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Fail(c, nil, "参数错误")
	}

	// 记录进数据库的用户上传图片文件路径 json
	sourceJson, err := util.SetJsonFields("[]", map[string]any{"0": sourcePath})
	if err != nil {
		response.Fail(c, nil, "参数错误")
		return
	}

	db := common.GetDB()

	// 计算要消耗的积分
	points := countPoints(width, height, 1)

	// 消耗积分
	pointRecord := model.Point{
		TaskId:       taskId,
		UID:          user.UID,
		BeforePoints: user.Points,
		Points:       points * -1,
		Tool:         "rembg",
		Remark:       "自动去除背景",
	}
	err = pointRecord.AddRecord(db, &user)
	if err != nil {
		if err.Error() == "don't have enough points" {
			response.Fail(c, nil, "积分不足")
		} else {
			response.Response(c, http.StatusInternalServerError, 500, nil, "积分使用失败")
		}
		return
	}

	// 创建任务并推送到消息队列
	task := &model.Task{
		TaskId:     taskId,
		UID:        user.UID,
		Tool:       "rembg",
		Params:     string(paramsJson),
		Source:     string(sourceJson),
		Output:     "[]",
		UsedPoints: points, // 根据生成图的宽高计算需要消耗的点数
		Status:     "pending",
		StartTime:  time.Now(),
	}

	err = createTask(task, &user, msgJson)
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		response.Response(c, http.StatusInternalServerError, 500, nil, "任务创建失败")
		return
	}

	response.Success(c, gin.H{"taskID": taskId}, "ok")
}

func createTask(task *model.Task, user *model.User, msgJson string) error {
	db := common.GetDB()

	// 记录进 reids 中
	err := addTaskToRedis(task.TaskId, user.UID)
	if err != nil {
		return errors.WithStack(err)
	}

	// 写入任务数据库
	if err := db.Create(task).Error; err != nil {
		delTaskToRedis(task.TaskId) // 失败从redis中回滚
		return errors.WithStack(err)
	}

	// 将消息发布到主题 "stabelDiffusion"
	err = common.Producer.Publish("stabelDiffusion", []byte(msgJson))
	if err != nil {
		delTaskToRedis(task.TaskId) // 失败从redis中回滚
		db.Model(task).Where("task_id = ?", task.TaskId).Update("status", "failed")
		return errors.WithStack(err)
	}

	return nil
}

func addTaskToRedis(taskId string, userID string) error {
	// 获取当前时间戳作为分数
	score := float64(time.Now().UnixNano())

	// 添加任务id到全局任务有序集合，按时间戳排序
	err := common.Rdb.ZAdd(common.Ctx, "tasks", &redis.Z{
		Score:  score,
		Member: taskId,
	}).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	// 添加任务到用户任务集合
	err = common.Rdb.SAdd(common.Ctx, userID, taskId).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	// 任务哈希Map
	err = common.Rdb.HMSet(common.Ctx, taskId, map[string]interface{}{
		"userID":   userID,
		"status":   "pending",
		"progress": 0,
		"endtime":  "",
		"output":   "",
		"params":   "",
	}).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func delTaskToRedis(taskId string) error {
	// 获取任务的 userID
	userID, err := common.Rdb.HGet(common.Ctx, taskId, "userID").Result()
	if err != nil {
		return nil
	}

	// 从排序集合中删除任务
	err = common.Rdb.ZRem(common.Ctx, "tasks", taskId).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	// 从用户集合中删除任务
	err = common.Rdb.SRem(common.Ctx, userID, taskId).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	// 删除任务详情记录
	err = common.Rdb.Del(common.Ctx, taskId).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// 保存上传图片并且返回图片的宽高
func saveUploadImage(imgBase64 string, savePath string) (string, int, int, error) {
	// 根据第一个逗号分割字符串
	parts := strings.SplitN(imgBase64, ",", 2)
	if len(parts) != 2 {
		return "", 0, 0, errors.WithStack(errors.New("base64 format error"))
	}

	// 判断图片种类
	var fileExt string
	if strings.Contains(parts[0], "image/jpeg") {
		fileExt = "jpg"
	} else if strings.Contains(parts[0], "image/png") {
		fileExt = "png"
	} else {
		return "", 0, 0, errors.WithStack(errors.New("image format error"))
	}

	// 解码base64字符串
	imgData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", 0, 0, errors.WithStack(err)
	}

	// 检查图片大小是否超过2MB
	maxUploadImageSize := viper.GetInt("image.maxUploadImageSize")
	if len(imgData) > maxUploadImageSize*1024*1024 {
		return "", 0, 0, errors.WithStack(errors.New("image size exceeds 2MB"))
	}

	// 解码图像数据以获取宽度和高度
	imgReader := strings.NewReader(string(imgData))
	img, _, err := image.Decode(imgReader)
	if err != nil {
		return "", 0, 0, errors.WithStack(err)
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// 计算文件 MD5 哈希
	md5Hash, err := util.Md5Hash(imgData)
	if err != nil {
		return "", 0, 0, err
	}

	// 根据哈希的前2位和最后2位，创建目录，减少文件目录中文件数量
	path1 := md5Hash[:2]  // 取哈希的前两位
	path2 := md5Hash[30:] // 取哈希的最后两位

	// 文件名和文件路径的创建
	saveFileName := md5Hash + "." + fileExt
	savePath = filepath.Join(savePath, path1, path2, saveFileName)
	fmt.Println(savePath)
	returnPath := filepath.Join(path1, path2, saveFileName)

	// 创建目录结构
	err = os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
	if err != nil {
		return "", 0, 0, errors.WithStack(err)
	}

	// 创建一个新的图片文件
	imgFile, err := os.Create(savePath)
	if err != nil {
		return "", 0, 0, errors.WithStack(err)
	}
	defer imgFile.Close()

	// 将解码后的数据写入图片文件
	_, err = imgFile.Write(imgData)
	if err != nil {
		return "", 0, 0, errors.WithStack(err)
	}

	return returnPath, width, height, nil
}

// 保存 stable diffusion 生成输出的图片
func saveOutputImage(imgBase64 string, savePath string) (string, error) {
	fileExt := "png"

	// 解码base64字符串
	imgData, err := base64.StdEncoding.DecodeString(imgBase64)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// 计算文件 MD5 哈希
	md5Hash, err := util.Md5Hash(imgData)
	if err != nil {
		return "", err
	}

	// 文件名和文件路径的创建
	saveFileName := md5Hash + "." + fileExt
	year := time.Now().Format("2006")
	month := time.Now().Format("01")
	day := time.Now().Format("02")
	// monthPath := year + "-" + month
	// dayPath := monthPath + "-" + day
	thumbPath := filepath.Join(savePath+"/thumb", year, month, day, saveFileName)
	savePath = filepath.Join(savePath, year, month, day, saveFileName)
	returnPath := filepath.Join(year, month, day, saveFileName)

	// 创建目录结构
	err = os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = os.MkdirAll(filepath.Dir(thumbPath), os.ModePerm)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// 创建一个新的图片文件
	imgFile, err := os.Create(savePath)
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer imgFile.Close()

	// 将解码后的数据写入图片文件
	_, err = imgFile.Write(imgData)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// 打开已保存的图片文件以生成缩略图
	imgFile.Seek(0, 0) // 将文件指针移到文件开始位置
	img, _, err := image.Decode(imgFile)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// 调用函数生成缩略图并保存
	err = generateThumb(img, thumbPath, "png", 400)
	if err != nil {
		return "", err
	}

	return returnPath, nil
}

// 生成等比例缩略图并保存为指定格式
func generateThumb(img image.Image, thumbPath string, format string, maxSize uint) error {
	// 获取原图的宽度和高度
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()

	// 创建文件
	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer thumbFile.Close()

	// 当不需要缩放的时候也会存储一份图片
	if originalHeight <= int(maxSize) && originalWidth <= int(maxSize) {
		switch format {
		case "png":
			err = png.Encode(thumbFile, img)
		case "jpg", "jpeg":
			// 质量设为 80，可以根据需求调整
			err = jpeg.Encode(thumbFile, img, &jpeg.Options{Quality: 80})
		default:
			return errors.New("unsupported image format: " + format)
		}
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	var newWidth, newHeight uint

	if originalWidth < originalHeight {
		newWidth = maxSize
		newHeight = uint(float64(originalHeight) * (float64(maxSize) / float64(originalWidth)))
	} else {
		newHeight = maxSize
		newWidth = uint(float64(originalWidth) * (float64(maxSize) / float64(originalHeight)))
	}

	// 生成等比例缩略图
	thumbImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

	// 根据格式保存缩略图
	switch format {
	case "png":
		err = png.Encode(thumbFile, thumbImg)
	case "jpg", "jpeg":
		// 质量设为 80，可以根据需求调整
		err = jpeg.Encode(thumbFile, thumbImg, &jpeg.Options{Quality: 80})
	default:
		return errors.New("unsupported image format: " + format)
	}

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
