package controller

import (
	"errors"
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/dto"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/util"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Regist(c *gin.Context) {
	DB := common.GetDB()

	// 获取参数
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm_password")
	tokenStr := c.PostForm("captcha_token")

	// 人机验证
	captchaToken, _, err := common.ParseToken(tokenStr)
	if err != nil || !captchaToken.Valid {
		response.Fail(c, nil, "人机验证失败")
		return
	}

	// 数据验证
	if !util.IsEmail(email) {
		response.Fail(c, nil, "邮箱格式错误")
		return
	}
	if len(password) < 6 {
		response.Fail(c, nil, "密码不能少于6位")
		return
	}
	if len(name) == 0 {
		name = util.RandomString(10)
	}

	// 判断邮箱是否已存在
	if isEmailExist(DB, email) {
		response.Fail(c, nil, "用户已存在")
		return
	}

	// 判断两次密码是否一样
	if password != confirmPassword {
		response.Fail(c, nil, "两次输入的密码不相等")
		return
	}

	// 创建用户
	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, nil, "加密错误")
		return
	}

	user := model.User{
		Name:     name,
		UID:      uuid.New().String(),
		Email:    email,
		Password: string(hashedPassword),
	}
	DB.Create(&user)

	// 注册成功 发放 token
	token, err := common.ReleaseToken(user)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error : %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "注册成功, 已自动登录")
}

func SignIn(c *gin.Context) {
	DB := common.GetDB()

	// 获取用户
	email := c.PostForm("email")
	password := c.PostForm("password")
	tokenStr := c.PostForm("captcha_token")

	// 人机验证
	captchaToken, _, err := common.ParseToken(tokenStr)
	if err != nil || !captchaToken.Valid {
		response.Fail(c, nil, "人机验证失败")
		return
	}

	// 数据验证
	if !util.IsEmail(email) {
		response.Fail(c, nil, "邮箱格式错误")
		return
	}
	if len(password) < 6 {
		response.Fail(c, nil, "密码不能少于6位")
		return
	}

	// 判断邮箱是否存在
	var user model.User
	DB.Where("email = ?", email).First(&user)
	if user.ID == 0 {
		response.Fail(c, nil, "用户不存在")
		return
	}

	// 判断密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		response.Fail(c, nil, "密码错误")
		return
	}

	// 发放token
	token, err := common.ReleaseToken(user)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error : %v", err)
		return
	}

	response.Success(c, gin.H{"token": token}, "欢迎登录")
}

func Info(c *gin.Context) {
	user, _ := c.Get("user")
	response.Success(c, gin.H{"user": dto.ToUserDto(user.(model.User))}, "")
}

func SetCaptcha(c *gin.Context) {
	DB := common.GetDB()

	// 获取用户
	email := c.PostForm("email")
	puzzleX := c.PostForm("puzzleX")

	// 数据验证
	if !util.IsEmail(email) {
		response.Fail(c, nil, "邮箱格式错误")
		return
	}

	puzzleXNum, err := strconv.Atoi(puzzleX)

	if err != nil {
		response.Fail(c, nil, "拼图数据上传错误")
		return
	}

	var userCaptcha model.UserCaptcha
	DB.Where("email = ?", email).First(&userCaptcha)

	if userCaptcha.Email == "" {
		userCaptcha = model.UserCaptcha{
			Email:   email,
			PuzzleX: puzzleXNum,
		}
		DB.Save(&userCaptcha)
	} else {
		// 邮箱存在，更新条数
		userCaptcha.PuzzleX = puzzleXNum
		userCaptcha.IsChecked = 0
		DB.Save(&userCaptcha)
	}

	response.Success(c, nil, "ok")
}

func ValidateCaptcha(c *gin.Context) {
	DB := common.GetDB()

	// 获取用户
	email := c.PostForm("email")
	puzzleX := c.PostForm("puzzleX")

	// 数据验证
	if !util.IsEmail(email) {
		response.Fail(c, nil, "邮箱格式错误")
		return
	}

	puzzleXNum, err := strconv.Atoi(puzzleX)

	if err != nil {
		response.Fail(c, nil, "拼图数据错误，请重试")
		return
	}

	var userCaptcha model.UserCaptcha
	result := DB.Where("email = ?", email).Where("is_checked = ?", 0).First(&userCaptcha)

	// 查找数据库对应的邮箱的puzzleX信息
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(c, nil, "操作频繁，请重试")
		return
	} else if result.Error != nil {
		response.Fail(c, nil, "数据库查询失败，请重试")
		return
	}

	// 更新拼图记录为已检查
	userCaptcha.IsChecked = 1
	DB.Save(&userCaptcha)

	// 判断UpdatedAt是否过期
	lifeTime := viper.GetInt("captcha.lifeTime")
	twoMinutesAgo := time.Now().Add(-time.Duration(lifeTime) * time.Minute)
	if userCaptcha.UpdatedAt.Before(twoMinutesAgo) {
		response.Fail(c, nil, "验证码已过期，请重试")
		return
	}

	// 验证 puzzleX 数据是否匹配
	offset := viper.GetInt("captcha.offset")
	if puzzleXNum <= userCaptcha.PuzzleX-offset || puzzleXNum >= userCaptcha.PuzzleX+offset {
		response.Fail(c, gin.H{"puzzleX": userCaptcha.PuzzleX, "puzzleNum": puzzleXNum}, "失败，请重试")
		return
	}

	token, err := common.ReleaseCaptchaToken(userCaptcha)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		log.Printf("token generate error : %v", err)
		return
	}

	// 成功响应
	response.Success(c, gin.H{"token": token}, "ok")
}

func isEmailExist(db *gorm.DB, email string) bool {
	var user model.User
	db.Where("email = ?", email).First(&user)

	return user.ID != 0
}
