package util

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

type ValidErrors []string

func (e ValidErrors) Error() string {
	if len(e) > 0 {
		return e[0]
	}
	return ""
}

func init() {
	// 注册自定义密码验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", PasswordValidator)
	}
}

// 错误转换机
func TranslateValidateError(err error, obj interface{}) ValidErrors {
	var errMsgs []string

	// 断言错误类型成功时进行错误的翻译
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		t := reflect.TypeOf(obj)

		for _, e := range validationErrors {
			field, _ := t.FieldByName(e.Field())
			label := field.Tag.Get("label")

			if label == "" {
				label = field.Name
			}

			// 检查字段是否是字符串类型
			isStringField := field.Type.Kind() == reflect.String

			switch e.Tag() {
			case "required":
				errMsgs = append(errMsgs, fmt.Sprintf("%s不能为空", label))
			case "min":

				if isStringField {
					errMsgs = append(errMsgs, fmt.Sprintf("%s长度不能小于%s个字符", label, e.Param()))
				} else {
					errMsgs = append(errMsgs, fmt.Sprintf("%s不能小于%s", label, e.Param()))
				}
			case "max":
				if isStringField {
					errMsgs = append(errMsgs, fmt.Sprintf("%s长度不能大于%s个字符", label, e.Param()))
				} else {
					errMsgs = append(errMsgs, fmt.Sprintf("%s不能大于%s", label, e.Param()))
				}
			case "len":
				if isStringField {
					errMsgs = append(errMsgs, fmt.Sprintf("%s长度必须为%s个字符", label, e.Param()))
				} else {
					errMsgs = append(errMsgs, fmt.Sprintf("%s必须等于%s", label, e.Param()))
				}
			case "email":
				errMsgs = append(errMsgs, fmt.Sprintf("%s格式不正确", label))
			case "eqfield":
				// 获取比较字段的名称
				compareFieldName := e.Param()
				compareField, _ := t.FieldByName(compareFieldName)
				compareLabel := compareField.Tag.Get("label")

				if compareLabel == "" {
					compareLabel = compareFieldName
				}
				errMsgs = append(errMsgs, fmt.Sprintf("%s与%s不匹配", label, compareLabel))
			case "password":
				errMsgs = append(errMsgs, fmt.Sprintf("%s必须包含至少8个字符，至少一个大写字母，一个小写字母和一个数字", label))
			case "numeric":
				errMsgs = append(errMsgs, fmt.Sprintf("%s必须是整数", label))
			default:
				errMsgs = append(errMsgs, fmt.Sprintf("%s校验错误", label))
			}
		}
		return errMsgs
	}

	// 未知错误记录进日志中
	LogErr(errors.WithStack(err), "./log/gin.log")
	errMsgs = append(errMsgs, "未知错误，请重试")

	return errMsgs
}

// 自定义密码验证器
func PasswordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// 检查密码长度是否至少为8个字符
	if len(password) < 8 {
		return false
	}

	// 检查是否包含至少一个大写字母
	hasUpper := false
	// 检查是否包含至少一个小写字母
	hasLower := false
	// 检查是否包含至少一个数字
	hasNumber := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		}
	}

	return hasUpper && hasLower && hasNumber
}
