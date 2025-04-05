package authservice

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"url_profile/internal/domain/models"
	"url_profile/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	CreateUser(email string, pass []byte) (*models.User, error)
}

type UserProvider interface {
	User(email string) (*models.User, error)
	UserById(id int) (*models.User, error)
	UpdateAboutMe(id int, text string) error
	AddLink(userID int, link models.ReqLink) error
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
	if _, err := a.userProvider.User(email); err != store.ErrUserNotFound {
		return http.StatusConflict, nil, fmt.Errorf("user with email %s already exists", email)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Info("error from bcrypt", slog.String("err", err.Error()))
		return http.StatusInternalServerError, nil, err
	}

	u, err := a.userSaver.CreateUser(email, hash)
	if err != nil {
		if errors.Is(err, store.ErrUserAlreadyExists) {
			return http.StatusConflict, nil, store.ErrUserAlreadyExists
		}

		if errors.Is(err, store.ErrUserRetrievalFailed) {
			return http.StatusInternalServerError, nil, store.ErrUserRetrievalFailed
		}

		return http.StatusInternalServerError, nil, store.ErrDatabaseOperation
	}

	return http.StatusCreated, u, nil
}

func (a *AuthService) User(email string) (*models.User, error) {
	u, err := a.userProvider.User(email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return nil, store.ErrUserNotFound
		}

		return nil, store.ErrDatabaseOperation
	}

	return u, nil
}

func (a *AuthService) UserById(id int) (*models.User, error) {
	u, err := a.userProvider.UserById(id)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return nil, store.ErrUserNotFound
		}

		return nil, store.ErrDatabaseOperation
	}

	return u, nil
}

func (a *AuthService) UpdateAboutMe(id int, text string) error {
	if err := a.userProvider.UpdateAboutMe(id, text); err != nil {
		return err
	}
	return nil
}

func (a *AuthService) AddLink(userID int, link models.ReqLink) error {
	if err := a.userProvider.AddLink(userID, link); err != nil {
		a.log.Debug("Failet to save link", slog.String("error", err.Error()))
		return err
	}

	return nil
}
