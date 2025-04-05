package sqlitestore

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"url_profile/internal/domain/models"
	"url_profile/internal/store"

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
	stmt, err := s.db.Prepare("INSERT INTO users (email, pass_hash) VALUES($1, $2) RETURNING id")
	if err != nil {
		s.log.Error("error from PREPARE SQL USERS", slog.String("err", err.Error()))
		return nil, store.ErrDatabaseOperation
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, pass)
	if err != nil {
		if isDuplicateKeyError(err) {
			s.log.Error("User Already Exists", slog.String("err", err.Error()))
			return nil, store.ErrUserAlreadyExists
		}
		s.log.Error("error from EXEC SQL USERS", slog.String("err", err.Error()))
		return nil, store.ErrDatabaseOperation
	}

	userID, err := res.LastInsertId()
	if err != nil {
		if exists, _ := s.User(email); exists != nil {
			return nil, store.ErrUserAlreadyExists
		}

		return nil, store.ErrDatabaseOperation
	}

	s.log.Debug("user created",
		slog.Int64("user_id", userID),
		slog.String("email", email),
	)

	u := &models.User{}
	if err := s.db.QueryRow("SELECT id, email FROM users WHERE id = ?", userID).Scan(
		&u.ID,
		&u.Email,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Error("newly created user not found - data inconsistency",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
			return nil, store.ErrUserRetrievalFailed
		}
		s.log.Error("failed to retrieve created user",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	s.log.Debug("successfully created and retrieved user",
		slog.Int("user_id", u.ID),
		slog.String("email", u.Email))
	return u, nil
}

func (s *Store) User(email string) (*models.User, error) {
	rows, err := s.db.Query(`
        SELECT 
            u.id, u.email, u.pass_hash, u.about_text,
            l.id, l.user_id, l.link_name, l.link_color,l.link_path
        FROM users u 
        LEFT JOIN links l ON u.id = l.user_id
        WHERE u.email = ?
    `, email)
	if err != nil {
		s.log.Error("failed to query user data",
			slog.String("email", email),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var user models.User
	var links []models.Link
	userFound := false

	for rows.Next() {
		var (
			link       models.Link
			linkID     sql.NullInt64
			linkUserID sql.NullInt64
			linkName   sql.NullString
			linkColor  sql.NullString
			linkPath   sql.NullString
		)

		if !userFound {
			err := rows.Scan(
				&user.ID, &user.Email, &user.HashedPassword, &user.AboutText,
				&linkID, &linkUserID, &linkName, &linkColor, &linkPath,
			)
			if err != nil {
				s.log.Error("failed to scan user data",
					slog.String("email", email),
					slog.String("error", err.Error()))
				return nil, fmt.Errorf("%w: failed to scan user data", store.ErrDataScanFailed)
			}
			userFound = true
		} else {
			var discardID int
			var discardEmail, discardAboutText string
			err := rows.Scan(
				&discardID, &discardEmail, &discardAboutText,
				&linkID, &linkUserID, &linkName, &linkColor, &linkPath,
			)
			if err != nil {
				s.log.Error("failed to scan link data",
					slog.String("email", email),
					slog.String("error", err.Error()))
				continue
			}
		}

		if linkID.Valid {
			link.ID = int(linkID.Int64)
			link.UserID = int(linkUserID.Int64)
			link.LinkName = linkName.String
			link.LinkColor = linkColor.String
			link.LinkPath = linkPath.String

			links = append(links, link)
		}
	}

	if err := rows.Err(); err != nil {
		s.log.Error("rows iteration error",
			slog.String("email", email),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: rows iteration failed", store.ErrDatabaseOperation)
	}

	if !userFound {
		s.log.Debug("user not found", slog.String("email", email))
		return nil, store.ErrUserNotFound
	}

	user.Links = links
	return &user, nil
}

func (s *Store) UserById(id int) (*models.User, error) {
	rows, err := s.db.Query(`
        SELECT 
            u.id, u.email, u.pass_hash, u.about_text,
            l.id, l.user_id, l.link_name, l.link_color, l.link_path
        FROM users u 
        LEFT JOIN links l ON u.id = l.user_id
        WHERE u.id = ?
    `, id)
	if err != nil {
		s.log.Error("failed to query user data",
			slog.Int("id", id),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}
	defer rows.Close()

	var user models.User
	var links []models.Link
	userFound := false

	for rows.Next() {
		var (
			link       models.Link
			linkID     sql.NullInt64
			linkUserID sql.NullInt64
			linkName   sql.NullString
			linkColor  sql.NullString
			linkPath   sql.NullString
		)

		if !userFound {
			err := rows.Scan(
				&user.ID, &user.Email, &user.HashedPassword, &user.AboutText,
				&linkID, &linkUserID, &linkName, &linkColor, &linkPath,
			)
			if err != nil {
				s.log.Error("failed to scan user data",
					slog.Int("id", id),
					slog.String("error", err.Error()))
				return nil, fmt.Errorf("%w: failed to scan user data", store.ErrDataScanFailed)
			}
			userFound = true
		} else {
			err := rows.Scan(
				new(interface{}), new(interface{}), new(interface{}), new(interface{}), // Игнорируем user fields
				&linkID, &linkUserID, &linkName, &linkColor, &linkPath,
			)
			if err != nil {
				s.log.Error("failed to scan link data",
					slog.Int("id", id),
					slog.String("error", err.Error()))
				continue
			}
		}

		if linkID.Valid {
			link = models.Link{
				ID:        int(linkID.Int64),
				UserID:    int(linkUserID.Int64),
				LinkName:  linkName.String,
				LinkColor: linkColor.String,
				LinkPath:  linkPath.String,
			}
			links = append(links, link)
		}
	}

	if err := rows.Err(); err != nil {
		s.log.Error("rows iteration error",
			slog.Int("id", id),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: rows iteration failed", store.ErrDatabaseOperation)
	}

	if !userFound {
		s.log.Debug("user not found", slog.Int("id", id))
		return nil, store.ErrUserNotFound
	}

	user.Links = links
	return &user, nil
}

func (s *Store) UpdateAboutMe(id int, text string) error {
	stmt, err := s.db.Prepare("UPDATE users SET about_text = ? WHERE id = ?")
	if err != nil {
		s.log.Error("failed to prepare update statement",
			slog.Int("user_id", id),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(text, id)
	if err != nil {
		s.log.Debug("error from EXEC SQL Update TextAbout", slog.String("err", err.Error()))
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		s.log.Error("failed to verify update",
			slog.Int("user_id", id),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: cannot verify update", store.ErrDatabaseOperation)
	}

	if rowsAffected == 0 {
		s.log.Warn("about text update affected 0 rows",
			slog.Int("user_id", id))
		return store.ErrNoRowsAffected
	}

	s.log.Debug("about text updated successfully",
		slog.Int("user_id", id))

	return nil
}

func (s *Store) AddLink(userID int, link models.ReqLink) error {
	stmt, err := s.db.Prepare("INSERT INTO links (user_id, link_name, link_color, link_path) VALUES (?, ?, ?, ?)")
	if err != nil {
		s.log.Error("failed to prepare update statement",
			slog.Int("user_id", userID),
			slog.Any("link", link))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	_, err = stmt.Exec(userID, link.LinkName, link.LinkColor, link.LinkPath)
	if err != nil {
		if isDuplicateKeyError(err) {
			s.log.Warn("duplicate link path",
				slog.Int("user_id", userID),
				slog.String("path", link.LinkPath))
			return store.ErrLinkAlreadyExists
		}

		if isForeignKeyError(err) {
			s.log.Warn("invalid user reference",
				slog.Int("user_id", userID))
			return store.ErrUserNotFound
		}

		s.log.Error("failed to insert link",
			slog.Int("user_id", userID),
			slog.Any("link", link),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	return nil
}

func (s *Store) UpdateLink(userID int, link *models.ReqUpdateLink) error {
	var exists bool

	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM links WHERE id = ? AND user_id = ?
		)`,
		link.LinkID, userID).Scan(&exists)

	if err != nil {
		s.log.Error("failed to check link existence",
			slog.Int("user_id", userID),
			slog.Int("link_id", link.LinkID),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	if !exists {
		s.log.Warn("link not found or doesn't belong to user",
			slog.Int("user_id", userID),
			slog.Int("link_id", link.LinkID))
		return store.ErrLinkNotFound
	}

	stmt, err := s.db.Prepare(`
        UPDATE links 
        SET 
            link_name = ?,
            link_color = ?,
            link_path = ?
        WHERE user_id = ? AND id = ?
    `)
	if err != nil {
		s.log.Error("failed to prepare update statement",
			slog.Int("user_id", userID),
			slog.Int("link_id", link.LinkID),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		link.LinkName,
		link.LinkColor,
		link.LinkPath,
		userID,
		link.LinkID,
	)
	if err != nil {
		s.log.Error("failed to execute link update",
			slog.Int("user_id", userID),
			slog.Int("link_id", link.LinkID),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	s.log.Debug("link updated successfully",
		slog.Int("user_id", userID),
		slog.Int("link_id", link.LinkID))
	return nil
}

func (s *Store) DeleteLink(userID int, linkID int) error {

	var exists bool
	err := s.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM links WHERE id = ? AND user_id = ?
		)`,
		linkID, userID).Scan(&exists)

	if err != nil {
		s.log.Error("failed to check link existence",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	if !exists {
		s.log.Warn("link not found or doesn't belong to user",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID))
		return store.ErrLinkNotFound
	}

	stmt, err := s.db.Prepare("DELETE FROM links WHERE id = ? AND user_id = ?")
	if err != nil {
		s.log.Error("failed to prepare delete statement",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(linkID, userID)
	if err != nil {
		s.log.Error("failed to execute delete link",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	return nil
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func isForeignKeyError(err error) bool {
	if err == nil {
		return false
	}
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "FOREIGN KEY constraint failed") ||
		strings.Contains(errorMsg, "SQLITE_CONSTRAINT_FOREIGNKEY") ||
		strings.Contains(errorMsg, "no such table:")
}
