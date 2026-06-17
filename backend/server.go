package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/sebasukodo/just-another-blog/backend/internal/handler"
)

type server struct {
	httpServer *http.Server
	handler    *handler.Handler
	cancel     context.CancelFunc
}

func newServer(h *handler.Handler, cancel context.CancelFunc) *server {
	mux := http.NewServeMux()

	s := &server{
		handler: h,
		cancel:  cancel,
	}

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%v", serverPort),
		Handler:           requestLogger(h.Logger)(mux),
		ReadHeaderTimeout: 15 * time.Second,
	}

	fileServerHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/static/", fileServerHandler)

	mux.HandleFunc("POST /api/users", h.RegisterUser)
	mux.HandleFunc("POST /api/users/login", h.LoginUser)
	mux.HandleFunc("GET /api/user", h.AuthMiddleware(h.CurrentUser))
	mux.HandleFunc("PUT /api/user", h.AuthMiddleware(h.UpdateUser))
	mux.HandleFunc("DELETE /api/user", h.AuthMiddleware(h.DeleteUser))

	mux.HandleFunc("GET /api/profiles/{username}", h.OptionalAuthMiddleware(h.GetProfile))
	mux.HandleFunc("POST /api/profiles/{username}/follow", h.AuthMiddleware(h.FollowUser))
	mux.HandleFunc("DELETE /api/profiles/{username}/follow", h.AuthMiddleware(h.UnfollowUser))

	mux.HandleFunc("GET /api/articles/{slug}", h.GetArticle)
	mux.HandleFunc("POST /api/articles", h.AuthMiddleware(h.CreateArticle))
	mux.HandleFunc("PUT /api/articles/{slug}", h.AuthMiddleware(h.UpdateArticle))
	mux.HandleFunc("DELETE /api/articles/{slug}", h.AuthMiddleware(h.DeleteArticle))

	mux.HandleFunc("GET /api/tags", h.GetTags)

	return s
}

func (s *server) start() error {
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}
	s.handler.Logger.Debug(fmt.Sprintf("server is running on port %v", serverPort))
	if err := s.httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *server) shutdown(ctx context.Context) error {
	s.handler.Logger.Debug("server is shutting down...")
	return s.httpServer.Shutdown(ctx)
}

func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			logger.Info(fmt.Sprintf("served request: %s %s", r.Method, r.URL.Path))
		})
	}
}
