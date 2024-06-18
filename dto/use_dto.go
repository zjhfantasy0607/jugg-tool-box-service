package dto

import "jugg-tool-box-service/model"

type UserDto struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Points int    `json:"points"`
}

func ToUserDto(user model.User) UserDto {
	return UserDto{
		Name:   user.Name,
		Email:  user.Email,
		Points: user.Points,
	}
}
