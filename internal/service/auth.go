package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type auth struct {
	log *slog.Logger
	tokenTTL time.Duration
	refreshTTL time.Duration
	dbProvider dbProvider
	tokenManager tokenManager
}

type dbProvider interface{
	SaveUser(ctx context.Context, email string, passHash []byte) (string, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	SaveRefreshToken(ctx context.Context, email string, refreshToken []byte, rttl time.Duration) (error)
}

type tokenManager interface{
	CreateJWT(user models.User, ttl time.Duration) (string, error)
	CreateRefresh() (string, error)
	ParseJwt(token string) (jwt.MapClaims, error)
}

func NewAuthService(log *slog.Logger, ttl time.Duration, rttl time.Duration, dbProvider dbProvider, tokenManager tokenManager) *auth {
	return &auth{
		log: log.With(slog.String("service", "auth")),
		tokenTTL: ttl,
		refreshTTL: rttl,
		dbProvider: dbProvider,
		tokenManager: tokenManager,
	}
}

func (a *auth) RegisterUser(ctx context.Context, email, password string) (string, error) {
	a.log.Info("attempt to register new user")
	a.log.Debug("got user email", slog.String("email", email))
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", slog.String("error", err.Error()))
		return "", fmt.Errorf("can't register user: %w", err)
	}
	id, err := a.dbProvider.SaveUser(ctx, email, passHash)
	if err != nil {
		a.log.Error("failed to save user", slog.String("error", err.Error()))
		return "", fmt.Errorf("can't register user: %w", err)
	}
	a.log.Info("user registered", slog.String("UserId", id))
	return id, nil
}

func (a *auth) Login(ctx context.Context, email, password string) (models.TokenPair, error) {
	a.log.Info("attempt to login user")
	a.log.Debug("got user email", slog.String("email", email))
	user, err := a.dbProvider.GetUserByEmail(ctx, email)
	if err != nil {
		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn(ErrInvalidCredentials.Error(), slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", ErrInvalidCredentials)
	}
	token, err := a.tokenManager.CreateJWT(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate access token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	refresh, err := a.tokenManager.CreateRefresh()
	if err != nil {
		a.log.Error("failed to generate refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	rfsHash, err := bcrypt.GenerateFromPassword([]byte(refresh), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate refresh token hash", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	err = a.dbProvider.SaveRefreshToken(ctx, email, rfsHash, a.refreshTTL)
	if err != nil {
		a.log.Error("failed to save refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	return models.TokenPair{
		AccessToken: token,
		RefreshToken: refresh,
	}, nil
}

func (a *auth) RefreshToken(ctx context.Context, access, refresh string) (models.TokenPair, error) {
	a.log.Info("attempt to refresh token pair")
	a.log.Debug("got tokens", slog.String("acces_token", access), slog.String("refresh_token", refresh))
	claims, err := a.tokenManager.ParseJwt(access)
	if err != nil {
		a.log.Info("not valid access token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", err)
	}
	email, ok := claims["email"]
	if !ok {
		a.log.Info("failed get email from token")
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", err)
	}
	user, err := a.dbProvider.GetUserByEmail(ctx, email.(string))
	if err != nil {
		a.log.Error("failed get user", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", err)
	}
	a.log.Debug("got user data", slog.Any("user", user))
	err = bcrypt.CompareHashAndPassword([]byte(user.RefreshHash), []byte(refresh))
	if err != nil {
		a.log.Info("not valid refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", err)
	}
	if time.Now().Unix() > user.ExpiresAt {
		a.log.Info("not valid refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", ErrValidRefresh)
	}
	token, err := a.tokenManager.CreateJWT(user, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate access token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	newRefresh, err := a.tokenManager.CreateRefresh()
	if err != nil {
		a.log.Error("failed to generate refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	rfsHash, err := bcrypt.GenerateFromPassword([]byte(newRefresh), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate refresh token hash", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	err = a.dbProvider.SaveRefreshToken(ctx, email.(string), rfsHash, a.refreshTTL)
	if err != nil {
		a.log.Error("failed to save refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	return models.TokenPair{
		AccessToken: token,
		RefreshToken: refresh,
	}, nil
}