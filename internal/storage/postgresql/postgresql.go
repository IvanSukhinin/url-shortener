package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"url-shortener/internal/config"
	"url-shortener/internal/storage"
)

const UniqueViolationErrCode = pq.ErrorCode("23505")

type Storage struct {
	db *sql.DB
}

type Alias struct {
	Id    uuid.UUID `json:"id"`
	Url   string    `json:"url"`
	Alias string    `json:"alias"`
}

func New(cfg config.Db) (*Storage, error) {
	const op = "storage.postgresql.New"
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Db)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveAlias(url string, alias string) error {
	const op = "storage.postgresql.SaveAlias"
	_, err := s.db.Query("INSERT INTO url_alias(url, alias) VALUES($1, $2)", url, alias)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == UniqueViolationErrCode {
			return storage.ErrAliasExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetUrl(alias string) (string, error) {
	const op = "storage.postgresql.GetUrl"
	row := s.db.QueryRow("SELECT url FROM url_alias WHERE alias = $1", alias)
	var resUrl string
	err := row.Scan(&resUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrAliasNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resUrl, nil
}

func (s *Storage) DeleteAlias(alias string) error {
	const op = "storage.postgresql.DeleteAlias"
	_, err := s.db.Query("DELETE FROM url_alias WHERE alias = $1", alias)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}
	return nil
}

func (s *Storage) GetAliasList() (*[]Alias, error) {
	const op = "storage.postgresql.GetAliasList"
	rows, err := s.db.Query("SELECT id, alias, url FROM url_alias")
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}
	defer rows.Close()
	var aliases []Alias

	for rows.Next() {
		var alias Alias
		if err := rows.Scan(&alias.Id, &alias.Alias, &alias.Url); err != nil {
			return &aliases, fmt.Errorf("%s, %w", op, err)
		}
		aliases = append(aliases, alias)
	}

	if err = rows.Err(); err != nil {
		return &aliases, fmt.Errorf("%s, %w", op, err)
	}

	return &aliases, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
