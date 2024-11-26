package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"song-lib/internal/config"
	"song-lib/internal/db/postgres"
	"song-lib/internal/domain"
	"song-lib/internal/integrations/songinfo"
	"song-lib/internal/repos"
	slogutils "song-lib/internal/utils/slog-utils"

	_ "song-lib/internal/controllers/v1"
	songcontroller "song-lib/internal/controllers/v1/song"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/otel"
)

const gracefulServerShutdownTimeout = time.Second * 5

var Tracer = otel.Tracer("gin-server")

//	@title			Song library
//	@version		1.0
//	@description	Library of song texts and metadata

// @BasePath	/api/v1
func Run(cfg config.Config) error {
	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(cfg.LogLevel))
	if err != nil {
		return errors.Wrap(err, "parse log level")
	}
	logger, err := newLogger(cfg.Env, logLevel)
	if err != nil {
		return err
	}
	slog.SetDefault(logger)

	migrator, err := migrate.New(
		"file://./deploy/migrations/postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.DBConfig.Username, cfg.DBConfig.Password,
			cfg.DBConfig.Host, cfg.DBConfig.Port,
			cfg.DBConfig.DBName, cfg.DBConfig.SSLMode))
	if err != nil {
		return errors.Wrap(err, "create migrator")
	}
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Wrap(err, "migrate")
	}

	postgresClient, err := postgres.NewClient(cfg.DBConfig)
	if err != nil {
		return errors.Wrap(err, "initialize Postgres client")
	}

	songRepository := repos.NewSongRepository(postgresClient)
	songInfoIntegration := songinfo.NewSongInfoIntegration(cfg.SongInfoIntegrationAPI)
	songService := domain.NewSongService(songRepository, songInfoIntegration)

	songController := songcontroller.NewSongController(songService)

	switch cfg.Env {
	case config.EnvLocal:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(setRequestIDMiddleware(), setLoggerMiddleware())
	engine.GET("api/v1/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	songController.RegisterRoutes(engine)

	srv := &http.Server{
		Addr:    cfg.HTTPServer.Host + ":" + cfg.HTTPServer.Port,
		Handler: engine.Handler(),
	}
	runServer(srv)

	return nil
}

func runServer(srv *http.Server) {
	slog.Info("starting server...")
	defer slog.Info("exited")

	serverExit := make(chan struct{}, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen:", "error", err)
			serverExit <- struct{}{}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), gracefulServerShutdownTimeout)
		defer cancel()
		slog.Info("shutting down server gracefully...", "timeout", gracefulServerShutdownTimeout.String())
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("graceful server shutdown:", "error", err)
		} else {
			slog.Info("server shut down gracefully")
		}
	case <-serverExit:
	}
}

func setRequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := ksuid.New().String()
		c.Set("requestID", traceID)

		c.Next()

		c.Header("X-Request-ID", traceID)
	}
}

func setLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default()
		if traceID, ok := c.Get("requestID"); ok {
			logger = logger.With("requestID", traceID.(string))
		}
		c.Set("logger", logger)

		c.Next()
	}
}

func newLogger(env config.Env, logLevel slog.Level) (logger *slog.Logger, err error) {
	switch env {
	case config.EnvLocal:
		logger = newPrettyLogger(logLevel)
	case config.EnvProd:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}),
		)
	default:
		return nil, errors.Wrap(err, "unknown env")
	}

	return logger, nil
}

func newPrettyLogger(logLevel slog.Level) *slog.Logger {
	opts := slogutils.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: logLevel,
		},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
