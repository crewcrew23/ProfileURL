package serviceinterface

import (
	"url_profile/internal/app/server/handlers/requestModel"
	"url_profile/internal/domain/models"
)

// TODO: create single interface
type UserService interface {
	CreateUser(email string, username string, password string) (code int, user *models.User, err error)
	User(email string) (*models.User, error)
	UserById(id int) (*models.User, error)
	UserByUsername(name string) (*models.User, error)
	UpdateAboutMe(id int, text string) error
	AddLink(userID int, link requestModel.ReqLink) error
	UpdateLink(userID int, link *requestModel.ReqUpdateLink) error
	DeleteLink(userID int, linkID int) error
}
