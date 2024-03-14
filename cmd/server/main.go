package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/app"
	c "github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/http/middleware"
	v1 "github.com/EwvwGeN/InHouseAd_assignment/internal/http/v1"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/jwt"
	l "github.com/EwvwGeN/InHouseAd_assignment/internal/logger"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/service"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/storage"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
}
func main() {
	flag.Parse()
	cfg, err := c.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("cant load config from path %s: %s", configPath, err.Error()))
	}
	logger := l.SetupLogger(cfg.LogLevel)
	logger.Info("logger is initiated")
	logger.Debug("config data", slog.Any("config", cfg))
	mainCtx, cancel := context.WithCancel(context.Background())

	jwtManager := jwt.NewJwtManager(cfg.SecretKey)

	postgres, err := storage.NewPostgresProvider(mainCtx, cfg.PostgresConfig)
	if err != nil {
		logger.Error("failed to get postgres provider", slog.String("error", err.Error()))
		os.Exit(1)
	}

	authService := service.NewAuthService(logger, cfg.TokenTTL, cfg.RefreshTTL, postgres, jwtManager)
	categoryService := service.NewCategoryService(logger, nil)

	hserver := app.NewHttpServer(cfg.HttpConfig, logger)
	hserver.RegisterHandler(
		"/api/register",
		v1.Register(logger, authService, cfg.Validator),
		http.MethodPost,
	)
	hserver.RegisterHandler(
		"/api/login",
		v1.Login(logger, authService),
		http.MethodPost,
	)
	hserver.RegisterHandler(
		"/api/refresh",
		v1.Refresh(logger, authService),
		http.MethodPost,
	)
	hserver.RegisterHandler(
		"/api/category/add",
		middleware.AuthMiddleware(logger, jwtManager, v1.CategoryAdd(logger, cfg.Validator, categoryService)),
		http.MethodPost,
	)
	hserver.RegisterHandler(
		"/api/category/{catCode}/edit",
		middleware.AuthMiddleware(logger, jwtManager, v1.CategoryEdit(logger, categoryService)),
		http.MethodPatch,
	)
	hserver.RegisterHandler(
		"/api/category/{catCode}/delete",
		middleware.AuthMiddleware(logger, jwtManager, v1.CategoryDelete(logger, categoryService)),
		http.MethodGet,
	)
	hserver.RegisterHandler(
		"/api/category/{catCode}",
		v1.CategoryGetOne(logger, categoryService),
		http.MethodGet,
	)
	hserver.RegisterHandler(
		"/api/categories",
		v1.CategoryGetAll(logger, categoryService),
		http.MethodGet,
	)
	logger.Info("loading end")
	errCh := hserver.RunServer(mainCtx)
	stopChecker := make(chan os.Signal, 1)
	signal.Notify(stopChecker, syscall.SIGTERM, syscall.SIGINT)
	<- stopChecker
	logger.Info("stopping service")
	cancel()
	err = <-errCh
	if err != nil {
		logger.Error("error while stopping http server", slog.String("error", err.Error()))
	}
	logger.Info("service stoped successfully")
}