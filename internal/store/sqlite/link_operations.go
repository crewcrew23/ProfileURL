package sqlitestore

import (
	"fmt"
	"log/slog"
	"url_profile/internal/app/server/http/handlers/requestModel"
	"url_profile/internal/store"
	errshandle "url_profile/internal/store/sqlite/errs"
	"url_profile/internal/store/sqlite/query"

	_ "github.com/mattn/go-sqlite3"
)

func (s *Store) insertLink(userID int, link requestModel.ReqLink) error {
	stmt, err := s.db.Prepare(query.InsertLink)
	if err != nil {
		s.log.Error("failed to prepare update statement",
			slog.Int("user_id", userID),
			slog.Any("link", link))
		return fmt.Errorf("%w: %v", store.ErrDatabaseOperation, err)
	}

	_, err = stmt.Exec(userID, link.LinkName, link.LinkColor, link.LinkPath)
	if err != nil {
		if errshandle.IsDuplicateKeyError(err) {
			s.log.Warn("duplicate link path",
				slog.Int("user_id", userID),
				slog.String("path", link.LinkPath))
			return store.ErrLinkAlreadyExists
		}

		if errshandle.IsForeignKeyError(err) {
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

func (s *Store) existsLink(userID int, linkID int) error {
	var exists bool

	err := s.db.QueryRow(query.ExistsLink,
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

	return nil
}

func (s *Store) updateLink(userID int, link *requestModel.ReqUpdateLink) error {
	stmt, err := s.db.Prepare(query.UpdateLink)
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

func (s *Store) deleteLink(userID int, linkID int) error {
	stmt, err := s.db.Prepare(query.DeleteLink)
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
