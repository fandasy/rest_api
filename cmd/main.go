package main

import (
	"context"
	"errors"
	"flag"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"restApi/internal/cache/rdb"
	"restApi/internal/config"
	image_storage "restApi/internal/files/image-storage"
	"restApi/internal/http-server/handlers/redirect"
	"restApi/internal/http-server/handlers/save"
	"restApi/internal/http-server/middleware/cors"
	"restApi/internal/http-server/middleware/logger"
	req_controller "restApi/internal/http-server/middleware/req-controller"
	"restApi/internal/lib/l"
	"restApi/internal/storage/psql"
	"restApi/pkg/e"
	"syscall"
	"time"
)

func main() {

	cfg, err := config.Load(mustGetConfigPath())
	if err != nil {
		panic(err)
	}

	log := l.SetupLogger(cfg.Env)

	storage, err := psql.New(cfg.StoragePath, log)
	if err != nil {
		panic(err)
	}

	if err := storage.Init(context.TODO()); err != nil {
		panic(err)
	}

	imgDir, err := image_storage.Init("img")
	if err != nil {
		panic(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cache, err := rdb.New(context.TODO(), cfg.Redis, log)
	if err != nil {
		panic(err)
	}

	rateLimiter := req_controller.New(cfg.ReqLimit)

	router := gin.New()

	router.Use(logger.Checking(log))
	router.Use(cors.Middleware()) // Для обработки OPTIONS запросов
	router.Use(gin.Recovery())
	router.Use(rateLimiter.Checking())

	router.POST("/url",
		save.New(context.TODO(),
			*log,
			storage,
			cfg.ImageSettings,
			imgDir,
		))

	router.GET("/id/:id",
		redirect.New(context.TODO(),
			*log,
			storage,
			cache,
			dir,
			imgDir,
		))

	log.Info("server starting", slog.String("address", cfg.HttpServer.Addr))

	srv := &http.Server{
		Addr:         cfg.HttpServer.Addr,
		Handler:      router,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Info("server closed")
			} else {
				panic(e.Wrap("failed to start http server", err))
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop
	log.Info("server stopping", slog.Any("signal", sign))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Info("server shutdown", slog.Any("error", err))
	}

	log.Info("server shutdown")
}

func mustGetConfigPath() string {
	var configPath string

	flag.StringVar(&configPath,
		"config",
		"",
		"config file path",
	)

	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath == "" {
			panic("config path is not specified")
		}
	}

	return configPath
}
