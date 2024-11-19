package psql

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"restApi/internal/storage"
	"restApi/pkg/e"

	_ "github.com/lib/pq"
)

type Storage struct {
	db  *sql.DB
	log *slog.Logger
}

func New(connStr string, log *slog.Logger) (*Storage, error) {
	const op = "storage.psql.New"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, e.Wrap(op, err)
	}

	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func (s *Storage) Save(ctx context.Context, imageUrl string, imageName string) (int, error) {
	const op = "storage.psql.Save"

	q := `INSERT INTO images (url, image_name) VALUES ($1, $2) RETURNING id`

	var id int

	if err := s.db.QueryRowContext(ctx, q, imageUrl, imageName).Scan(&id); err != nil {
		return -1, e.Wrap(op, err)
	}

	return id, nil
}

func (s *Storage) Get(ctx context.Context, ID int) (string, error) {
	const op = "storage.psql.Get"

	q := `SELECT image_name FROM images WHERE id = $1`

	var imgPath string

	if err := s.db.QueryRowContext(ctx, q, ID).Scan(&imgPath); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", e.Wrap(op, storage.ErrURLNotFound)
		}

		return "", e.Wrap(op, err)
	}

	return imgPath, nil
}

func (s *Storage) IsExists(ctx context.Context, imageName string) (bool, error) {
	const op = "storage.psql.IsExists"

	q := `SELECT EXISTS(SELECT 1 FROM images WHERE image_name = $1)`

	var exists bool

	if err := s.db.QueryRowContext(ctx, q, imageName).Scan(&exists); err != nil {
		return false, e.Wrap(op, err)
	}

	return exists, nil
}

func (s *Storage) Init(ctx context.Context) error {
	const op = "storage.psql.Init"

	q := `CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    image_name TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_id ON images (id);`

	_, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return e.Wrap(op, err)
	}

	return nil
}
