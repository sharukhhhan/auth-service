package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"medods-tz/config"
	v1 "medods-tz/internal/controller/http/v1"
	"medods-tz/internal/repository"
	"medods-tz/internal/sender"
	"medods-tz/internal/service"
	"medods-tz/pkg/logger"
	"medods-tz/pkg/validator"
	"net/http"
	"os"
	"os/signal"
)

func Run(configPath string) {
	ctx := context.Background()

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	err = logger.SetupLogrus(cfg.Log.Level, cfg.Log.LogPath)
	if err != nil {
		log.Fatal(fmt.Errorf("error while initializing server logs: %w", err))
	}
	scrLogs, err := logger.NewFileLogger(fmt.Sprintf("%s/security.log", cfg.Log.LogPath))
	if err != nil {
		log.Fatal(fmt.Errorf("error while initializing logs for security: %w", err))
	}

	log.Debug("Initializing smtp-client")
	sender := sender.NewSender(sender.NewEmailSender(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password))

	err = sender.EnsureSMTPConnection()
	if err != nil {
		log.Errorf("error while connecting to smtp-client: %w", err)
	}

	log.Debug("Connecting postgres...")
	pgURL := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.Name)
	pg, err := pgx.Connect(ctx, pgURL)
	if err != nil {
		log.Fatal(fmt.Errorf("error connecting postgres: %w", err))
	}
	defer pg.Close(ctx)

	log.Debug("Running migrations...")
	err = RunMigrations(pgURL, cfg.Database.Postgres.MigrationPath)
	if err != nil {
		log.Debug(fmt.Errorf("error running migrations: %w", err))
	}

	log.Debug("Initializing repositories...")
	repositories := repository.NewRepository(pg)

	log.Debug("Initializing services")
	dependencies := service.ServicesDependencies{
		Repository:      repositories,
		TokenTTL:        cfg.JWT.TokenTTL,
		SignKey:         cfg.JWT.SignKey,
		RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
		SecurityLog:     scrLogs,
		Sender:          sender,
	}
	services := service.NewService(dependencies)

	log.Debug("Initializing handlers and routes...")
	handler := echo.New()
	handler.Validator = validator.NewCustomValidator()
	v1.NewRouter(handler, services, cfg.Log.LogPath)

	log.Info("Starting http server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)

	httpServer := &http.Server{
		Addr:    cfg.HTTP.Port,
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	done := make(chan struct{})

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error starting server: %v", err)
		}
	}()

	// Graceful Shutdown
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			log.Fatalf("error shutdown: %v", err)
		}

		close(done)
	}()

	<-done
	log.Info("Server stopped gracefully.")
}
