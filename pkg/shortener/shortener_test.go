package shortener

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name  string
		count int64
	}{

		{
			name:  "generate more short keys",
			count: 1000000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := make(map[string]int, tt.count)

			var i int64

			for i < tt.count {
				shortKey := Generate(i)

				if _, ok := buf[shortKey]; ok {
					t.Error("generate shortener fail: ", shortKey)
					break
				}
				buf[shortKey]++

				i++
			}
		})
	}
}
