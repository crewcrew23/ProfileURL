package sqlitestore

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"url_profile/internal/domain/models"
	"url_profile/internal/store"
	errshandle "url_profile/internal/store/sqlite/errs"
	"url_profile/internal/store/sqlite/query"

	_ "github.com/mattn/go-sqlite3"
)

func (s *Store) insertUser(email string, pass []byte) (int64, error) {

	stmt, err := s.db.Prepare(query.InsertUser)
	if err != nil {
		s.log.Error("error from PREPARE SQL USERS", slog.String("err", err.Error()))
		return 0, store.ErrDatabaseOperation
	}
	defer stmt.Close()

	res, err := stmt.Exec(email, pass)
	if err != nil {
		if errshandle.IsDuplicateKeyError(err) {
			s.log.Error("User Already Exists", slog.String("err", err.Error()))
			return 0, store.ErrUserAlreadyExists
		}
		s.log.Error("error from EXEC SQL USERS", slog.String("err", err.Error()))
		return 0, store.ErrDatabaseOperation
	}

	userID, err := res.LastInsertId()
	if err != nil {
		if exists, _ := s.User(email); exists != nil {
			return 0, store.ErrUserAlreadyExists
		}
		return 0, store.ErrDatabaseOperation
	}

	s.log.Debug("user created",
		slog.Int64("user_id", userID),
		slog.String("email", email))

	return userID, nil
}

func (s *Store) createdUser(userID int64) (*models.User, error) {
	u := &models.User{}
	err := s.db.QueryRow(query.CreatedUser, userID).Scan(
		&u.ID,
		&u.Email,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Error("newly created user not found - data inconsistency",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()))
			return nil, store.ErrUserRetrievalFailed
		}

		s.log.Error("failed to retrieve created user",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	return u, nil
}

func (s *Store) userRowsByEmail(email string) (*sql.Rows, error) {

	rows, err := s.db.Query(query.UsersRowsByEmail, email)
	if err != nil {
		s.log.Error("failed to query user data",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	return rows, nil
}

func (s *Store) userRowsByID(id int) (*sql.Rows, error) {
	rows, err := s.db.Query(query.UsersRowsByID, id)
	if err != nil {
		s.log.Error("failed to query user data",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	return rows, nil
}

func (s *Store) scanUserRows(rows *sql.Rows) (*models.User, error) {
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
					slog.String("error", err.Error()))
				return nil, fmt.Errorf("%w: failed to scan user data", store.ErrDataScanFailed)
			}
			userFound = true
		} else {
			var discardID int
			var discardEmail, discardAboutText string
			err := rows.Scan(
				&discardID, &discardEmail, &user.HashedPassword, &discardAboutText,
				&linkID, &linkUserID, &linkName, &linkColor, &linkPath,
			)
			if err != nil {
				s.log.Error("failed to scan link data",
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
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: rows iteration failed", store.ErrDatabaseOperation)
	}

	s.log.Debug("User on DB", slog.Any("data:", user))
	if !userFound {
		s.log.Debug("user not found")
		return nil, store.ErrUserNotFound
	}

	user.Links = links
	return &user, nil
}

func (s *Store) updateAboutMe(id int, text string) (sql.Result, error) {
	stmt, err := s.db.Prepare(query.UpdateAboutMe)
	if err != nil {
		s.log.Error("failed to prepare update statement",
			slog.Int("user_id", id),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(text, id)
	if err != nil {
		s.log.Debug("error from EXEC SQL Update TextAbout", slog.String("err", err.Error()))
		return nil, err
	}

	return res, nil
}

func (s *Store) rowsAffectedCheck(res sql.Result) error {
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		s.log.Error("failed to verify update",
			slog.String("error", err.Error()))
		return fmt.Errorf("%w: cannot verify update", store.ErrDatabaseOperation)
	}

	if rowsAffected == 0 {
		s.log.Warn("about text update affected 0 rows")
		return store.ErrNoRowsAffected
	}

	s.log.Debug("about text updated successfully")

	return nil
}
