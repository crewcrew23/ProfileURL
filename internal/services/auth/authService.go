package authservice

import (
	"database/sql"
	"log/slog"
	"net/http"
	"url_profile/internal/domain/models"

	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	CreateUser(email string, pass []byte) error
}

type UserProvider interface {
	User(email string) (*models.User, error)
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

func (a *AuthService) CreateUser(email string, password string) (int, error) {
	if _, err := a.userProvider.User(email); err != sql.ErrNoRows {
		return http.StatusConflict, err // TODO: проработаь ошибки
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Info("error from bcrypt", slog.String("err", err.Error()))
		return http.StatusInternalServerError, err // TODO: проработаь ошибки
	}

	if err := a.userSaver.CreateUser(email, hash); err != nil {
		a.log.Info("error from Create User", slog.String("err", err.Error()))
		return http.StatusInternalServerError, err // TODO: проработаь ошибки
	}

	return http.StatusCreated, nil
}

func (a *AuthService) User(email string) (*models.User, error) {
	u, err := a.userProvider.User(email)
	if err != nil {
		return nil, err
	}

	return u, nil
}
