package cache

import (
	"context"
	"errors"
)

type Cache interface {
	Set(ctx context.Context, k string, v string) error
	Get(ctx context.Context, k string) (string, error)
}

var (
	ErrKeyNotFound = errors.New("key not found")
)
