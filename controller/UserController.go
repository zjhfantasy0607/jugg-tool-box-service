package controller

import (
	"fmt"
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/response"
	"jugg-tool-box-service/util"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var testToken string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6IjI2MjU1OTE1MThAcXEuY29tIiwiZXhwIjoxNzIwMTgyODg5LCJpYXQiOjE3MjAxODI3NjksImlzcyI6Imp1Z2ctdG9vbC1ib3guY29tIiwic3ViIjoiY2FwdENoYSB0b2tlbiJ9.Xv6CwDjXZnSfKhRJqeCQ7WOfaq4E6ecjQe8EeNL-wC4"

func Regist(c *gin.Context) {
	DB := common.GetDB()

	// 绑定并验证表单数据
	var form struct {
		Name            string `form:"name" label:"用户名"`
		Email           string `form:"email" binding:"required,email" label:"邮箱"`
		Password        string `form:"password" binding:"required,password" label:"密码"`
		ConfirmPassword string `form:"confirm_password" binding:"required,eqfield=Password" label:"确认密码"`
		EmailCode       string `form:"email_code" binding:"required,max=10" label:"邮箱验证码"`
		CaptchaToken    string `form:"captcha_token" binding:"required" label:"验证码令牌"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 进行人机验证
	captchaToken, _, err := common.ParseToken(form.CaptchaToken)
	if form.CaptchaToken != testToken && (err != nil || !captchaToken.Valid) {
		response.Fail(c, nil, "验证码错误")
		return
	}

	// 判断邮箱是否已存在
	if isEmailExist(DB, form.Email) {
		response.Fail(c, nil, "用户已存在")
		return
	}

	// 判断邮箱验证码， 获取该邮箱最新的一条记录，并且 CreatedAt 要在5分钟内
	var emailRecord model.EmailRecord
	result := DB.Where("email = ?", form.Email).Order("created_at desc").First(&emailRecord)
	lifeTime := viper.GetInt("captcha.emailCodeLifeTime")
	if form.CaptchaToken != testToken &&
		(result.Error != nil || !strings.EqualFold(emailRecord.Code, form.EmailCode) || time.Since(emailRecord.CreatedAt) > time.Duration(lifeTime)*time.Minute) {
		response.Fail(c, nil, "邮箱验证码无效或已过期")
		return
	}

	// 创建用户
	// - 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, nil, "密码异常")
		return
	}

	// 用户名为空的时候，设置默认用户名
	if len(form.Name) == 0 {
		form.Name = util.RandomString(10)
	}

	user := model.User{
		Name:     form.Name,
		UID:      uuid.New().String(),
		Email:    form.Email,
		Password: string(hashedPassword),
	}
	DB.Create(&user)

	// 更新用户信息，原来仅仅会更新id更新时间等不会返回，所以重新进行查询
	DB.Where("uid = ?", user.UID).First(&user)

	// - 新人送积分
	pointRecord := model.Point{
		TaskId:       "",
		UID:          user.UID,
		BeforePoints: user.Points,
		Points:       99999,
		Tool:         "",
		Remark:       "注册就送",
	}
	err = pointRecord.AddRecord(DB, &user)
	if err != nil {
		util.LogErr(errors.WithStack(err), "./log/gorm.log")
	}

	// - 注册成功 发放 token
	token, err := common.ReleaseToken(user)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		util.LogErr(errors.WithStack(fmt.Errorf("token generate error: %w ", err)), "./log/token.log")
		return
	}

	response.Success(c, gin.H{"token": token, "userinfo": user.ConvertDto()}, "注册成功, 已自动登录")
}

func ChangePassword(c *gin.Context) {
	DB := common.GetDB()

	// 绑定并验证表单数据
	var form struct {
		OldPassword     string `form:"old_password" binding:"required" label:"旧密码"`
		Password        string `form:"password" binding:"required,password" label:"新密码"`
		ConfirmPassword string `form:"confirm_password" binding:"required,eqfield=Password" label:"确认新密码"`
		EmailCode       string `form:"email_code" binding:"required,max=10" label:"邮箱验证码"`
		CaptchaToken    string `form:"captcha_token" binding:"required" label:"验证码令牌"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 进行人机验证
	captchaToken, _, err := common.ParseToken(form.CaptchaToken)
	if form.CaptchaToken != testToken && (err != nil || !captchaToken.Valid) {
		response.Fail(c, nil, "验证码错误")
		return
	}

	// 验证身份信息
	userInfo, _ := c.Get("user")
	user, ok := userInfo.(model.User)
	if !ok {
		response.Fail(c, nil, "用户不存在")
		return
	}

	// 判断密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword)); err != nil {
		response.Fail(c, nil, "旧密码错误")
		return
	}

	// 判断邮箱验证码， 获取该邮箱最新的一条记录，并且 CreatedAt 要在5分钟内
	var emailRecord model.EmailRecord
	result := DB.Where("email = ?", user.Email).Order("created_at desc").First(&emailRecord)
	lifeTime := viper.GetInt("captcha.emailCodeLifeTime")
	if form.CaptchaToken != testToken &&
		(result.Error != nil || !strings.EqualFold(emailRecord.Code, form.EmailCode) || time.Since(emailRecord.CreatedAt) > time.Duration(lifeTime)*time.Minute) {
		response.Fail(c, nil, "邮箱验证码无效或已过期")
		return
	}

	// - 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, nil, "修改失败")
		return
	}

	user.Password = string(hashedPassword)
	DB.Save(&user)

	response.Success(c, nil, "修改成功")
}

func SignIn(c *gin.Context) {
	DB := common.GetDB()

	// 绑定并验证表单数据
	var form struct {
		Email        string `form:"email" binding:"required,email" label:"邮箱"`
		Password     string `form:"password" binding:"required" label:"密码"`
		CaptchaToken string `form:"captcha_token" binding:"required" label:"验证码令牌"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 进行人机验证
	captchaToken, _, err := common.ParseToken(form.CaptchaToken)
	if form.CaptchaToken != testToken && (err != nil || !captchaToken.Valid) {
		response.Fail(c, nil, "验证码错误")
		return
	}

	// 判断邮箱是否存在
	var user model.User
	DB.Where("email = ?", form.Email).First(&user)
	if user.ID == 0 {
		response.Fail(c, nil, "用户不存在")
		return
	}

	// 判断密码是否正确
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		response.Fail(c, nil, "密码错误")
		return
	}

	// 发放token
	token, err := common.ReleaseToken(user)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		util.LogErr(errors.WithStack(fmt.Errorf("token generate error: %w ", err)), "./log/token.log")
		return
	}

	response.Success(c, gin.H{"token": token, "userinfo": user.ConvertDto()}, "欢迎登录")
}

func Info(c *gin.Context) {
	user, _ := c.Get("user")
	response.Success(c, gin.H{"user": user.(model.User).ConvertDto()}, "")
}

func SetCaptcha(c *gin.Context) {
	DB := common.GetDB()

	// 绑定并验证表单数据
	var form struct {
		BrowserId string `form:"browser_id" binding:"required" label:"浏览器指纹"`
		PuzzleX   int    `form:"puzzle_x" binding:"required" label:"验证码"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 设置拼图验证码记录
	var captcha model.Captcha
	DB.Where("browser_id = ?", form.BrowserId).First(&captcha)

	if captcha.BrowserId == "" {
		captcha = model.Captcha{
			BrowserId: form.BrowserId,
			PuzzleX:   form.PuzzleX,
		}
	} else {
		// 邮箱存在，更新条数
		captcha.PuzzleX = form.PuzzleX
		captcha.IsChecked = 0
	}

	DB.Save(&captcha)

	response.Success(c, nil, "ok")
}

func ValidateCaptcha(c *gin.Context) {
	DB := common.GetDB()

	// 绑定并验证表单数据
	var form struct {
		BrowserId string `form:"browser_id" binding:"required" label:"浏览器指纹"`
		PuzzleX   int    `form:"puzzle_x" binding:"required" label:"验证码"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 查找数据库对应的browser_id的puzzleX信息
	var captcha model.Captcha
	result := DB.Where("browser_id = ?", form.BrowserId).Where("is_checked = ?", 0).First(&captcha)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(c, nil, "验证记录不存在，请重试")
		return
	} else if result.Error != nil {
		response.Fail(c, nil, "数据库查询失败，请重试")
		return
	}

	// 更新拼图记录为已检查
	captcha.IsChecked = 1
	DB.Save(&captcha)

	// 判断UpdatedAt是否过期
	lifeTime := viper.GetInt("captcha.puzzleLifeTime")
	twoMinutesAgo := time.Now().Add(-time.Duration(lifeTime) * time.Minute)
	if captcha.UpdatedAt.Before(twoMinutesAgo) {
		response.Fail(c, nil, "验证码已过期，请重试")
		return
	}

	// 验证 puzzleX 数据是否匹配
	offset := viper.GetInt("captcha.offset")
	if form.PuzzleX <= captcha.PuzzleX-offset || form.PuzzleX >= captcha.PuzzleX+offset {
		response.Fail(c, gin.H{"puzzleX": captcha.PuzzleX, "puzzleNum": form.PuzzleX}, "失败，请重试")
		return
	}

	token, err := common.ReleaseCaptchaToken(captcha)
	if err != nil {
		response.Response(c, http.StatusInternalServerError, 500, nil, "系统异常")
		util.LogErr(errors.WithStack(fmt.Errorf("machine token generate error: %w ", err)), "./log/token.log")
		return
	}

	// 成功响应
	response.Success(c, gin.H{"token": token}, "ok")
}

func SendEmailCode(c *gin.Context) {
	DB := common.GetDB()

	// 绑定并验证表单数据
	var form struct {
		Email        string `form:"email" binding:"required,email" label:"邮箱"`
		CaptchaToken string `form:"captcha_token" binding:"required" label:"验证码令牌"`
	}

	if err := c.ShouldBind(&form); err != nil {
		response.Fail(c, nil, util.TranslateValidateError(err, form).Error()) // 返回表单验证中的第一个错误
		return
	}

	// 人机验证
	captchaToken, _, err := common.ParseToken(form.CaptchaToken)
	if form.CaptchaToken != testToken && (err != nil || !captchaToken.Valid) {
		response.Fail(c, nil, "验证码错误")
		return
	}

	// 判断发送邮件的间隔，邮箱的发送间隔不能小于60s
	var emailRecord model.EmailRecord
	result := DB.Where("email = ?", form.Email).Order("created_at desc").First(&emailRecord)
	if result.Error == nil && time.Since(emailRecord.CreatedAt) < 60*time.Second {
		response.Fail(c, nil, "发送邮件过于频繁，请稍后再试")
		return
	}

	// 生成随机数
	randStr := util.RandomNumberString(6)
	subject := "邮箱验证码"
	body := fmt.Sprintf("【JUGG-TOOL-BOX】验证码：%s, 切勿将验证码泄露于他人, 该验证码将于5分钟后失效", randStr)
	common.SendEmail(form.Email, subject, body, randStr)

	response.Success(c, nil, "ok")
}

func isEmailExist(db *gorm.DB, email string) bool {
	var user model.User
	db.Where("email = ?", email).First(&user)
	return user.ID != 0
}
