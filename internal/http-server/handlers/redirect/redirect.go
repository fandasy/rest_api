package redirect

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"path"
	"restApi/internal/cache"
	"restApi/internal/lib/l/sl"
	"strconv"

	resp "restApi/internal/lib/api/response"

	"restApi/internal/storage"
)

func New(ctx context.Context, log slog.Logger, db storage.Storage, cdb cache.Cache, dir string, imgDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.redirect.New"

		log := &log

		clientIP := c.ClientIP()

		log = log.With(
			slog.String("op", op),
			slog.String("client_ip", clientIP),
		)

		idStr := c.Param("id")
		if idStr == "" {
			log.Info("id is empty")

			c.JSON(http.StatusOK, resp.Error("not found"))

			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Info("invalid id")

			c.JSON(http.StatusOK, resp.Error("invalid id"))

			return
		}

		imageName, err := cdb.Get(ctx, idStr)
		if err != nil {
			if errors.Is(err, cache.ErrKeyNotFound) {
				log.Debug("key not found")

				goto get_from_database
			}

			log.Error("failed to get image", slog.String("id", idStr), sl.Err(err))

			goto get_from_database
		}

		goto redirect_to_image

	get_from_database:
		imageName, err = db.Get(ctx, id)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("id not found", slog.String("id", idStr))

				c.JSON(http.StatusOK, resp.Error("not found"))

				return
			}

			log.Error("failed to get image", slog.String("id", idStr), sl.Err(err))

			c.JSON(http.StatusOK, resp.Error("internal error"))

			return
		}

		err = cdb.Set(ctx, idStr, imageName)
		if err != nil {
			log.Error("failed to set image", slog.String("id", idStr), sl.Err(err))
		}

	redirect_to_image:
		log.Debug("image found", slog.String("id", idStr))

		filepath := path.Join(dir, imgDir, imageName)

		http.ServeFile(c.Writer, c.Request, filepath)
	}
}
