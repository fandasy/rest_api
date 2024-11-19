package storage

import (
	"context"
	"errors"
)

type Storage interface {
	Save(ctx context.Context, imageUrl string, imageName string) (int, error)
	Get(ctx context.Context, ID int) (string, error)
	IsExists(ctx context.Context, imageName string) (bool, error)
}

var (
	ErrURLNotFound           = errors.New("URL not found")
	ErrShortenedNameNotFound = errors.New("shortened name not found")
)
