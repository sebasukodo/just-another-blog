package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

type server struct {
	httpServer *http.Server
	logger     *slog.Logger
	cancel     context.CancelFunc
}

func newServer(logger *slog.Logger, cancel context.CancelFunc) *server {
	mux := http.NewServeMux()

	s := &server{
		logger: logger,
		cancel: cancel,
	}

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%v", serverPort),
		Handler: requestLogger(logger)(mux),
	}

	fileServerHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/static/", fileServerHandler)

	return s
}

func (s *server) start() error {
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.logger.Debug(fmt.Sprintf("Server is running on port %v", serverPort))
	if err := s.httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *server) shutdown(ctx context.Context) error {
	s.logger.Debug("Server is shutting down...")
	return s.httpServer.Shutdown(ctx)
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			logger.Info(fmt.Sprintf("Served request: %s %s", r.Method, r.URL.Path))
		})
	}
}
