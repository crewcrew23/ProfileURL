package sqlitestore

import (
	"database/sql"
	"log/slog"
	"url_profile/internal/app/server/handlers/requestModel"
	"url_profile/internal/domain/models"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db  *sql.DB
	log *slog.Logger
}

func New(dbPath string, log *slog.Logger) *Store {

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)

	}

	if err := conn.Ping(); err != nil {
		panic(err)
	}

	return &Store{
		db:  conn,
		log: log,
	}
}

func (s *Store) CreateUser(email string, username string, pass []byte) (*models.User, error) {
	userID, err := s.insertUser(email, username, pass)
	if err != nil {
		return nil, err
	}

	u, err := s.createdUser(userID)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) User(email string) (*models.User, error) {
	s.log.Debug("Email is DB layer:", slog.String("email", email))
	rows, err := s.userRowsByEmail(email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := s.scanUserRows(rows)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) UserById(id int) (*models.User, error) {
	rows, err := s.userRowsByID(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := s.scanUserRows(rows)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) UserByUsername(name string) (*models.User, error) {
	rows, err := s.userRowsByUsername(name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u, err := s.scanUserRows(rows)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) UpdateAboutMe(id int, text string) error {
	res, err := s.updateAboutMe(id, text)
	if err != nil {
		return err
	}

	if err := s.rowsAffectedCheck(res); err != nil {
		return err
	}

	return nil
}

func (s *Store) AddLink(userID int, link requestModel.ReqLink) error {
	if err := s.insertLink(userID, link); err != nil {
		return err
	}

	return nil
}

func (s *Store) UpdateLink(userID int, link *requestModel.ReqUpdateLink) error {

	if err := s.existsLink(userID, link.LinkID); err != nil {
		return err
	}

	if err := s.updateLink(userID, link); err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteLink(userID int, linkID int) error {

	if err := s.existsLink(userID, linkID); err != nil {
		return err
	}

	if err := s.deleteLink(userID, linkID); err != nil {
		return err
	}

	return nil
}
