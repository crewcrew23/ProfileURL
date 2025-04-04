package sqlitestore

import (
	"database/sql"
	"fmt"
	"log/slog"
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

func (s *Store) CreateUser(email string, pass []byte) (*models.User, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	stmt, err := tx.Prepare("INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id")
	if err != nil {
		s.log.Debug("error from PREPARE SQL USERS", slog.String("err", err.Error()))
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, pass)
	if err != nil {
		s.log.Debug("error from EXEC SQL USERS", slog.String("err", err.Error()))
		return nil, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	linkStmt, err := tx.Prepare("INSERT INTO links (user_id, link_name, link_color, link_path) VALUES(?, ?, ?, ?)")
	if err != nil {
		s.log.Debug("error from PREPARE SQL LINKS", slog.String("err", err.Error()))
		return nil, err
	}
	defer linkStmt.Close()

	_, err = linkStmt.Exec(userID, "", "", "")
	if err != nil {
		s.log.Debug("error from EXEC SQL LINKS", slog.String("err", err.Error()))
		return nil, err
	}

	s.log.Debug("user created with default link",
		slog.Int64("user_id", userID),
		slog.String("email", email),
	)

	u := &models.User{}
	if err := tx.QueryRow("SELECT id, email FROM users WHERE id = ?", userID).Scan(
		&u.ID,
		&u.Email,
	); err != nil {
		s.log.Debug("error from EXEC SQL FIND USER", slog.String("err", err.Error()))
		return nil, err
	}

	s.log.Debug("Created in store:", slog.Any("user:", u))
	return u, nil
}

func (s *Store) User(email string) (*models.User, error) {
	rows, err := s.db.Query(`
        SELECT 
            u.id, u.email, pass_hash, u.about_text,
            l.link_name, l.link_color, l.link_path
        FROM users u
        LEFT JOIN links l ON u.id = l.user_id
        WHERE u.email = ?
    `, email)
	if err != nil {
		return nil, fmt.Errorf("failed to query user with links: %w", err)
	}
	defer rows.Close()

	var user models.User
	var links []models.Link
	userFound := false

	for rows.Next() {
		var link models.Link
		if !userFound {
			err := rows.Scan(
				&user.ID, &user.Email, &user.HashedPassword, &user.AboutText,
				&link.LinkName, &link.LinkColor, &link.LinkPath,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan user: %w", err)
			}
			userFound = true
		} else {
			err := rows.Scan(nil, nil, nil, nil, &link.LinkName, &link.LinkColor, &link.LinkPath)
			if err != nil {

			}
		}

		if link.LinkName != "" || link.LinkPath != "" {
			links = append(links, link)
		}
	}

	if !userFound {
		return nil, sql.ErrNoRows
	}

	user.Links = links
	return &user, nil

}
func (s *Store) UserById(id int) (*models.User, error) {
	rows, err := s.db.Query(`
        SELECT 
            u.id, u.email, u.about_text,
            l.link_name, l.link_color, l.link_path
        FROM users u
        LEFT JOIN links l ON u.id = l.user_id
        WHERE u.id = ?
    `, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query user with links: %w", err)
	}
	defer rows.Close()

	var user models.User
	var links []models.Link
	userFound := false

	for rows.Next() {
		var link models.Link
		if !userFound {
			err := rows.Scan(
				&user.ID, &user.Email, &user.AboutText,
				&link.LinkName, &link.LinkColor, &link.LinkPath,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan user: %w", err)
			}
			userFound = true
		} else {
			err := rows.Scan(nil, nil, nil, nil, &link.LinkName, &link.LinkColor, &link.LinkPath)
			if err != nil {

			}
		}

		if link.LinkName != "" || link.LinkPath != "" {
			links = append(links, link)
		}
	}

	if !userFound {
		return nil, sql.ErrNoRows
	}

	user.Links = links
	return &user, nil

}
