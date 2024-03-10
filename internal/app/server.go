package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	v1 "github.com/EwvwGeN/InHouseAd_assignment/internal/http/v1"
	"github.com/gorilla/mux"
)

type server struct {
	cfg config.HttpConfig
	log *slog.Logger
	router *mux.Router
}

func NewHttpServer(cfg config.HttpConfig, log *slog.Logger) *server {
	return &server{
		cfg: cfg,
		log: log,
		router: mux.NewRouter(),
	}
}

func (s *server) RunServer(ctx context.Context) (errCloseCh chan error) {
	s.log.Info("starting server")
	s.configureRouter()
	errCloseCh = make(chan error)
	srv := &http.Server{
		Handler: s.router,
		Addr:    fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	s.log.Info("starting listening", slog.String("addres", srv.Addr))
	go func() {
		<-ctx.Done()
		s.log.Info("Graceful shutdown server")
		errCloseCh <- srv.Shutdown(context.Background())
	}()
	go srv.ListenAndServe()
	return
}

func(s *server) RegisterHandler(url string, handler http.HandlerFunc, method string) {
	s.router.HandleFunc(
		url,
		handler,
	).Methods(method)
}

func (s *server) configureRouter() {
	s.router.HandleFunc(
		"/api/healthcheck",
		v1.Healthcheck(s.cfg.PingTimeout)).
	Methods(http.MethodGet)
}

