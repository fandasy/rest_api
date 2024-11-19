package img_compressor

import (
	"context"
	"errors"
	"image/jpeg"
	"os"
	"testing"
)

const (
	validUrl_1 = "https://ir.ozone.ru/s3/multimedia-7/c1000/6755179327.jpg"
	validUrl_2 = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRVNyUgiDjGP2BXtaLCP48USXG0l9sttdYNgw&s"

	maxWidth  = 1000
	maxHeight = 1000

	chars = "@%#*+=:~-. "
)

func TestGet(t *testing.T) {
	type args struct {
		ctx                 context.Context
		url                 string
		reductionPercentage float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errStr  error
	}{
		{
			name: "Valid URL with 50% reduction",
			args: args{
				ctx:                 context.Background(),
				url:                 validUrl_1,
				reductionPercentage: 0.5,
			},
			wantErr: false,
		},
		{
			name: "Valid URL with 0% reduction",
			args: args{
				ctx:                 context.Background(),
				url:                 validUrl_2,
				reductionPercentage: 0.0,
			},
			wantErr: false,
		},
		{
			name: "Invalid URL",
			args: args{
				ctx:                 context.Background(),
				url:                 "https://example.com/notValidURL",
				reductionPercentage: 0.0,
			},
			wantErr: true,
			errStr:  ErrPageNotFound,
		},
		{
			name: "Invalid image format",
			args: args{
				ctx:                 context.Background(),
				url:                 "https://img.goodfon.ru/wallpaper/big/2/94/anime-devushka-cepi.webp",
				reductionPercentage: 0.0,
			},
			wantErr: true,
			errStr:  ErrIncorrectFormat,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			img, err := Get(tt.args.ctx, tt.args.url, tt.args.reductionPercentage, maxWidth, maxHeight, chars)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {

				fileName := tt.name + ".jpg"
				file, err := os.Create(fileName)
				if err != nil {
					t.Errorf("failed to create file %s: %v", fileName, err)
					return
				}
				defer file.Close()

				if err := jpeg.Encode(file, img, nil); err != nil {
					t.Errorf("failed to encode image to JPEG: %v", err)
				}

			} else {

				if !errors.Is(err, tt.errStr) {
					t.Errorf("Get() error = %v, errStr %v", err, tt.errStr)
				}

			}
		})
	}
}
