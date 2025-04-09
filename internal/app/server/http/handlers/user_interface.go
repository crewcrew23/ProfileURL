package handler

import (
	"url_profile/internal/app/server/http/handlers/requestModel"
	"url_profile/internal/domain/models"
)

// TODO: need sigle interface
type UserService interface {
	CreateUser(model *requestModel.SignUpModel) (code int, user *models.User, err error)
	User(email string) (*models.User, error)
	UserByUsername(name string) (*models.User, error)
	UserById(id int) (*models.User, error)
	UpdateAboutMe(id int, text string) error
	AddLink(userID int, link requestModel.ReqLink) error
	UpdateLink(userID int, link *requestModel.ReqUpdateLink) error
	DeleteLink(userID int, linkID int) error
}
