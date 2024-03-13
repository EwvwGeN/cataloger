package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/domain/models"
	"github.com/EwvwGeN/InHouseAd_assignment/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type auth struct {
	log *slog.Logger
	tokenTTL time.Duration
	refreshTTL time.Duration
	dbRepo dbRepo
	tokenManager tokenManager
}

type dbRepo interface{
	SaveUser(ctx context.Context, email string, passHash string) (error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	SaveRefreshToken(ctx context.Context, email string, refreshToken string, rttl int64) (error)
}

type tokenManager interface{
	CreateJWT(user models.User, ttl time.Duration) (string, error)
	CreateRefresh() (string, error)
	MustParseJwt(token string) (map[string]interface{}, error)
}

func NewAuthService(log *slog.Logger, ttl time.Duration, rttl time.Duration, dbRepo dbRepo, tokenManager tokenManager) *auth {
	return &auth{
		log: log.With(slog.String("service", "auth")),
		tokenTTL: ttl,
		refreshTTL: rttl,
		dbRepo: dbRepo,
		tokenManager: tokenManager,
	}
}

func (a *auth) RegisterUser(ctx context.Context, email, password string) (error) {
	a.log.Info("attempt to register new user")
	a.log.Debug("got user email", slog.String("email", email))
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("failed to generate password hash", slog.String("error", err.Error()))
		return fmt.Errorf("can't register user: %w", err)
	}
	err = a.dbRepo.SaveUser(ctx, email, string(passHash))
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			a.log.Warn("failed to save user", slog.String("error", err.Error()))
			return ErrUserExist
		}
		a.log.Error("failed to save user", slog.String("error", err.Error()))
		return fmt.Errorf("can't register user: %w", err)
	}
	a.log.Info("user registered")
	return nil
}

func (a *auth) Login(ctx context.Context, email, password string) (models.TokenPair, error) {
	a.log.Info("attempt to login user")
	a.log.Debug("got user email", slog.String("email", email))
	user, err := a.dbRepo.GetUserByEmail(ctx, email)
	if err != nil {
		a.log.Error("failed to get user", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PassHash), []byte(password)); err != nil {
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
	err = a.dbRepo.SaveRefreshToken(ctx, email, string(rfsHash), time.Now().Add(a.refreshTTL).Unix())
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
	claims, err := a.tokenManager.MustParseJwt(access)
	if err != nil {
		a.log.Info("not valid access token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", err)
	}
	email, ok := claims["email"]
	if !ok {
		a.log.Info("failed get email from token")
		return models.TokenPair{}, fmt.Errorf("failed refresh token pair: %w", err)
	}
	user, err := a.dbRepo.GetUserByEmail(ctx, email.(string))
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
		a.log.Info("refresh token live expired")
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
	err = a.dbRepo.SaveRefreshToken(ctx, email.(string), string(rfsHash), time.Now().Add(a.refreshTTL).Unix())
	if err != nil {
		a.log.Error("failed to save refresh token", slog.String("error", err.Error()))
		return models.TokenPair{}, fmt.Errorf("can't login user: %w", err)
	}
	return models.TokenPair{
		AccessToken: token,
		RefreshToken: newRefresh,
	}, nil
}