package authservice

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"url_profile/internal/domain/models"

	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	CreateUser(email string, pass []byte) (*models.User, error)
}

type UserProvider interface {
	User(email string) (*models.User, error)
	UserById(id int) (*models.User, error)
}

type AuthService struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
}

func New(log *slog.Logger, userSaver UserSaver, userProvider UserProvider) *AuthService {
	return &AuthService{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
	}
}

func (a *AuthService) CreateUser(email string, password string) (int, *models.User, error) {
	if _, err := a.userProvider.User(email); err != sql.ErrNoRows {
		return http.StatusConflict, nil, fmt.Errorf("user with email %s already exists", email) // TODO: проработаь ошибки
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Info("error from bcrypt", slog.String("err", err.Error()))
		return http.StatusInternalServerError, nil, err // TODO: проработаь ошибки
	}

	u, err := a.userSaver.CreateUser(email, hash)
	if err != nil {
		a.log.Info("error from Create User", slog.String("err", err.Error()))
		return http.StatusInternalServerError, nil, err // TODO: проработаь ошибки
	}

	return http.StatusCreated, u, nil
}

func (a *AuthService) User(email string) (*models.User, error) {
	u, err := a.userProvider.User(email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (a *AuthService) UserById(id int) (*models.User, error) {
	u, err := a.userProvider.UserById(id)
	if err != nil {
		return nil, err
	}

	return u, nil
}
