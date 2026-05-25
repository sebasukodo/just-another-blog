package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" //sql driver
	"github.com/pressly/goose/v3"
	"github.com/sebasukodo/just-another-blog/backend/internal/config"
)

const filepathRoot = "./static"
const dbPort = "5432"
const serverPort = "7337"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	status := run(ctx, cancel)
	cancel()
	os.Exit(status)
}

func run(ctx context.Context, cancel context.CancelFunc) int {

	godotenv.Load()
	cfg := config.Load()

	logger, closeLogger, err := initializeLogger(cfg.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		return 1
	}
	defer func() {
		if err := closeLogger(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close logger: %v\n", err)
		}
	}()

	databaseURL := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", cfg.DBUser, cfg.DBPassword, cfg.DBHost, dbPort, cfg.DBName)

	logger.Debug("Connecting to database...")
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not connect to database at %v: %v", databaseURL, err))
		return 1
	}
	logger.Debug("Connected to database.")

	logger.Debug("Running migrations...")
	if err := goose.Up(db, "./sql/schema"); err != nil {
		logger.Error(fmt.Sprintf("Migration failed: %v", err))
		return 1
	}
	logger.Debug("Migrations completed successfully!")

	s := newServer(logger, cancel)
	var serverErr error
	go func() {
		serverErr = s.start()
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.shutdown(shutdownCtx); err != nil {
		logger.Info(fmt.Sprintf("Failed to shutdown server: %v", err))
		return 1
	}

	if serverErr != nil {
		logger.Info(fmt.Sprintf("Server error: %v", serverErr))
		return 1
	}

	return 0
}

type closeFunc func() error

func initializeLogger(logFile string) (*slog.Logger, closeFunc, error) {
	handlers := []slog.Handler{
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}
	closers := []closeFunc{}

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}
		bufferedFile := bufio.NewWriterSize(file, 8192)

		close := func() error {
			if err := bufferedFile.Flush(); err != nil {
				return fmt.Errorf("failed to flush log file: %w", err)
			}
			if err = file.Close(); err != nil {
				return fmt.Errorf("failed to close log file: %w", err)
			}
			return nil
		}
		handlers = append(handlers, slog.NewTextHandler(bufferedFile, &slog.HandlerOptions{Level: slog.LevelInfo}))
		closers = append(closers, close)
	}

	closer := func() error {
		var errs []error
		for _, close := range closers {
			if err := close(); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}
	return slog.New(slog.NewMultiHandler(handlers...)), closer, nil
}
