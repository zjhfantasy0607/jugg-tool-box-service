package common

import (
	"errors"
	"jugg-tool-box-service/util"
	"log"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

func SendEmail(targetEmail string, subject string, body string) error {
	// 配置SMTP服务器信息
	smtpHost := viper.GetString("smtp.host")
	smtpPort := viper.GetInt("smtp.port")
	smtpUsername := viper.GetString("smtp.username")
	smtpPassword := viper.GetString("smtp.password")

	log.Println(viper.GetString("smtp.host"))
	log.Println(viper.GetString("smtp.port"))
	log.Println(viper.GetString("smtp.username"))
	log.Println(viper.GetString("smtp.password"))
	log.Println(viper.GetString("smtp.from"))

	// 检查邮箱地址是否正确
	if !util.IsEmail(targetEmail) {
		return errors.New("invalid email address")
	}

	// 配置邮件信息
	fromName := viper.GetString("smtp.fromName")
	fromEmail := viper.GetString("smtp.fromEmail")

	// 创建新的邮件消息
	m := gomail.NewMessage()
	m.SetHeader("From", fromName+" <"+fromEmail+">")
	m.SetHeader("To", targetEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	// 配置SMTP拨号器
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUsername, smtpPassword)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		log.Fatalf("发送邮件失败 %s : %v", targetEmail, err)
	}

	log.Println("邮件发送成功!")

	return nil
}
