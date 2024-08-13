package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3" // init sqlite3 driver
)

// реализация для sqlite

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	// константа нужна чтобы показывать место возникновения ошибки для дебага
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	//

	// создаём таблицу если она не существует(миграции не предусмотрены)
	// url-shortener будет брать длинную ссылку, и делать из неё короткую, заменяя длинную часть на alias
	// Prepare возвращает подготовленный запрос
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS ind_alias on url(alias)
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// выполняем запрос
	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// int64 - это индекс созданной записи
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	// вставляем новую запись(новый url)
	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES(?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// проверка на ошибку, что введён алиас, который уже существует
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	// возвращаем id
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

// получаем url
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s:prepare statement: %w", op, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s:execute statement: %w", op, err)
	}

	return resURL, nil
}

// удаляем url
func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.Delete"

	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s:prepare statement: %w", op, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s:exec statement: %w", op, err)
	}
	return nil
}
