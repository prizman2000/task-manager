package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"lebedev.vr/task-manager/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS task(
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			name TEXT NOT NULL UNIQUE,
			created_at DATE DEFAULT (datetime('now','localtime')),
			status NUMERIC DEFAULT 0
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveTask(taskName string) (int64, error) {
	const op = "storage.sqlite.SaveTask"

	stmt, err := s.db.Prepare(`INSERT INTO task(name) VALUES(?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(taskName)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrTaskExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetTask(taskId int64) (string, error) {
	const op = "storage.sqlite.GetTask"

	stmt, err := s.db.Prepare(`SELECT name FROM task WHERE id = ?`)
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resName string
	err = stmt.QueryRow(taskId).Scan(&resName)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrTaskNotFound
		}
		
		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resName, nil
}