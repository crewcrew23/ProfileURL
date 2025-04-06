package handler

import "url_profile/internal/domain/models"

type UserService interface {
	CreateUser(email string, password string) (code int, user *models.User, err error)
	User(email string) (*models.User, error)
	UserById(id int) (*models.User, error)
	UpdateAboutMe(id int, text string) error
	AddLink(userID int, link models.ReqLink) error
	UpdateLink(userID int, link *models.ReqUpdateLink) error
	DeleteLink(userID int, linkID int) error
}
