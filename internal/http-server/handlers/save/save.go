package save

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"image/jpeg"
	"log/slog"
	"net/http"
	"os"
	"path"
	"restApi/pkg/shortener"

	"restApi/internal/config"
	img_compressor "restApi/pkg/img-compressor"

	"restApi/internal/lib/l/sl"
	"restApi/internal/storage"
	"restApi/pkg/validate"

	resp "restApi/internal/lib/api/response"
)

type Request struct {
	URL                   string  `json:"url"`
	CompressionPercentage float64 `json:"compression_percentage,omitempty"`
}

type Response struct {
	resp.Response
	ID int `json:"id,omitempty"`
}

func New(ctx context.Context, log slog.Logger, db storage.Storage, imgCfg config.ImageSettings, imgDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.save.New"

		log := &log

		clientIP := c.ClientIP()

		log = log.With(
			slog.String("op", op),
			slog.String("client_ip", clientIP),
		)

		var req Request

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Error("failed to decode request", sl.Err(err))

			c.JSON(http.StatusOK, resp.Error("failed to decode request"))

			return
		}

		log.Debug("request body decoded", slog.Any("request", req))

		if ok := validate.URL(req.URL); !ok {
			log.Info("invalid url", slog.String("url", req.URL))

			c.JSON(http.StatusOK, resp.Error("invalid url"))

			return
		}

		asciiImage, err := img_compressor.Get(ctx, req.URL, req.CompressionPercentage, imgCfg.MaxWidth, imgCfg.MaxHeight, imgCfg.Chars)
		if err != nil {
			if errors.Is(err, img_compressor.ErrPageNotFound) {
				log.Info("page not found", slog.String("url", req.URL))

				c.JSON(http.StatusOK, resp.Error("page not found"))

				return

			} else if errors.Is(err, img_compressor.ErrIncorrectFormat) {
				log.Info(err.Error(), slog.String("url", req.URL))

				c.JSON(http.StatusOK, resp.Error(err.Error()))

				return
			}

			log.Error("failed to get image", sl.Err(err))

			c.JSON(http.StatusOK, resp.Error("internal error"))

			return
		}

		var name string

		for {
			name = shortener.Generate(int64(asciiImage.Stride))

			isExists, err := db.IsExists(ctx, name)
			if err != nil {
				log.Error("failed to check existence", sl.Err(err))

				c.JSON(http.StatusOK, resp.Error("internal error"))

				return
			}

			if !isExists {
				break
			}
		}

		filename := name + ".jpg"
		file, err := os.Create(path.Join(imgDir, filename))
		if err != nil {
			log.Error("failed to create file", sl.Err(err))

			c.JSON(http.StatusOK, resp.Error("internal error"))

			return
		}

		defer file.Close()

		id, err := db.Save(ctx, req.URL, filename)
		if err != nil {
			log.Error("failed to save url", slog.String("url", req.URL), sl.Err(err))

			c.JSON(http.StatusOK, resp.Error("internal error"))

			return
		}

		if err := jpeg.Encode(file, asciiImage, nil); err != nil {
			log.Error("failed to encode image", sl.Err(err))

			c.JSON(http.StatusOK, resp.Error("internal error"))

			return
		}

		log.Debug("url saved", slog.String("url", req.URL))

		c.JSON(http.StatusOK, Response{
			Response: resp.OK(),
			ID:       id,
		})
	}
}
