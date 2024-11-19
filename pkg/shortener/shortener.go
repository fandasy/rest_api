package shortener

import (
	"math/rand"
	"time"
)

const (
	charset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLength = 10
)

func Generate(seed int64) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + seed))

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[r.Intn(len(charset))]
	}

	return string(shortKey)
}
