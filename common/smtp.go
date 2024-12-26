package common

import (
	"fmt"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/util"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func SendEmail(targetEmail string, subject string, body string, remark string) error {
	// 配置smtp服务器信息
	smtpHost := viper.GetString("smtp.host")
	smtpPort := viper.GetInt("smtp.port")
	smtpUsername := viper.GetString("smtp.username")
	smtpPassword := viper.GetString("smtp.password")

	// 配置邮件信息
	fromName := viper.GetString("smtp.fromName")
	fromEmail := viper.GetString("smtp.fromEmail")

	// 创建新的邮件消息
	m := gomail.NewMessage()
	m.SetHeader("From", fromName+" <"+fromEmail+">")
	m.SetHeader("To", targetEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// 配置smtp拨号器
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)
	if smtpPort == 465 {
		d.SSL = true
	}

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		util.LogErr(errors.WithStack(fmt.Errorf("发送邮件失败 %s : %w", targetEmail, err)), "./log/email.log")
		return err
	}

	DB := GetDB()

	// 将成功的邮件记录存进数据库
	userCaptcha := model.EmailRecord{
		Email:   targetEmail,
		Subject: subject,
		Body:    body,
		Code:    remark,
	}
	DB.Save(&userCaptcha)

	return nil
}
