package common

import (
	"jugg-tool-box-service/model"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
)

var jwtKey = []byte("sz-drvtry_crect")

type Claims struct {
	UID string
	jwt.StandardClaims
}

type CaptchaClaims struct {
	Email string
	jwt.StandardClaims
}

func ReleaseToken(user model.User) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UID: user.UID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "jugg-tool-box.com",
			Subject:   "user token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ReleaseCaptchaToken(userCaptcha model.UserCaptcha) (string, error) {
	lifeTime := viper.GetInt("captcha.lifeTime")
	expirationTime := time.Now().Add(time.Duration(lifeTime) * time.Minute)
	claims := &CaptchaClaims{
		Email: userCaptcha.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "jugg-tool-box.com",
			Subject:   "captCha token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseToken(tokenString string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtKey, nil
	})

	return token, claims, err
}
