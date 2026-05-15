package httparticle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	ssomiddleware "sso/internal/http/sso"
	articleservice "sso/internal/services/article"
	"time"

	"github.com/gorilla/mux"
)

type ArticleServer struct {
	log    *slog.Logger
	server *http.Server
	router *mux.Router
}

func NewArticleServer(log *slog.Logger, address string, timeout time.Duration, handler *articleservice.ArticleHandler, secret string) *ArticleServer {
	router := mux.NewRouter()

	router.HandleFunc("/", handler.GetRoute).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/articles/{id:[0-9]+}", handler.GetArticleById).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/articles/authors/{id:[0-9]+}", handler.GetArticleByAuthor).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/auth/login", handler.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/register", handler.Register).Methods(http.MethodPost)

	protected := router.PathPrefix("/api/v1").Subrouter()

	protectHandler := ssomiddleware.AuthMiddleware(secret)
	protected.Use(protectHandler)

	protected.HandleFunc("/articles", handler.Create).Methods(http.MethodPost)
	protected.HandleFunc("/articles/{id:[0-9]+}", handler.Edit).Methods(http.MethodPut)

	httpServer := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	return &ArticleServer{
		log:    log,
		server: httpServer,
		router: router,
	}
}

func (s *ArticleServer) Run() error {
	const op = "http.server.Run"
	s.log.Info("starting http server", slog.String("addr", s.server.Addr))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *ArticleServer) Stop(ctx context.Context) error {
	const op = "http.server.Stop"
	s.log.Info("stopping http server", slog.String("addr", s.server.Addr))

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
