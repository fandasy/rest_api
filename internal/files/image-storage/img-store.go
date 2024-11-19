package image_storage

import (
	"os"
	"restApi/pkg/e"
)

func Init(name string) (string, error) {

	if _, err := os.Stat(name); os.IsNotExist(err) {
		if err := os.Mkdir(name, 0774); err != nil {
			return "", e.Wrap("can't create a images dir", err)
		}
	}

	return name, nil
}
