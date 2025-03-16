package v1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/EwvwGeN/cataloger/internal/config"
	"github.com/EwvwGeN/cataloger/internal/domain/httpmodels"
	"github.com/EwvwGeN/cataloger/internal/domain/models"
	v1 "github.com/EwvwGeN/cataloger/internal/http/v1"
	"github.com/EwvwGeN/cataloger/internal/jwt"
	"github.com/EwvwGeN/cataloger/internal/service"
	"github.com/EwvwGeN/cataloger/internal/service/mocks"
	"github.com/EwvwGeN/cataloger/internal/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type usrTestSuite struct {
	suite.Suite
	cfg config.Config
	userRepoMock *mocks.UserRepo
	registerHandler http.HandlerFunc
	loginHandler http.HandlerFunc
	refreshHanlder http.HandlerFunc
	jwtManager interface{
		ParseJwt(token string) (map[string]interface{}, error)
		MustParseJwt(token string) (map[string]interface{}, error)
		CreateJWT(user models.User, ttl time.Duration) (string, error)
		CreateRefresh() (string, error)
	}
}

func TestUserSuiteRun(t *testing.T) {
	suite.Run(t, new(usrTestSuite))
}

func (suite *usrTestSuite) SetupSuite() {
	cfg := config.Config{
		Validator: config.Validator{
			EmailValidate: `(\w+@\w+\.\w+)`,
			PasswordValidate: `.{5,}`,
		},
		TokenTTL: time.Duration(10)*time.Minute,
		RefreshTTL: time.Duration(10)*time.Minute,
		SecretKey: "test_key",
	}
	suite.userRepoMock = mocks.NewUserRepo(suite.T())
	tokenMng := jwt.NewJwtManager(cfg.SecretKey)
	lg := slog.New(
		slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelError}),
	)
	authService := service.NewAuthService(lg, cfg.TokenTTL, cfg.RefreshTTL, suite.userRepoMock, tokenMng)
	regHandler := v1.Register(lg, authService, cfg.Validator)
	logHandler := v1.Login(lg, authService)
	refHandler := v1.Refresh(lg, authService)
	suite.cfg = cfg
	suite.jwtManager = tokenMng
	suite.registerHandler = regHandler
	suite.loginHandler = logHandler
	suite.refreshHanlder = refHandler
}

func (suite *usrTestSuite) Test_Register() {
	registered := make(map[string]struct{})
	tests := []struct{
		name string
		req httpmodels.RegisterRequest
		dbQuery bool
		wantCode int
		wantSave bool
	}{
		{
			name: "happy_pass",
			req: httpmodels.RegisterRequest{
				Email: "test@test.test",
				Password: "12345",
			},
			dbQuery: true,
			wantCode: http.StatusCreated,
			wantSave: true,
		},
		{
			name: "duplicate_user",
			req: httpmodels.RegisterRequest{
				Email: "test@test.test",
				Password: "12345",
			},
			dbQuery: true,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_email",
			req: httpmodels.RegisterRequest{
				Email: "t",
				Password: "12345",
			},
			dbQuery: false,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "not_valid_password",
			req: httpmodels.RegisterRequest{
				Email: "test@test.test",
				Password: "1",
			},
			dbQuery: false,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		err := json.NewEncoder(&jsonBody).Encode(&tt.req)
		suite.Require().NoError(err, "failed to encode request")
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/user/register", &jsonBody)
		if tt.dbQuery {
			suite.userRepoMock.On("SaveUser", mock.Anything, tt.req.Email, mock.AnythingOfType("string")).Once().
			Return(func(ctx context.Context, email string, passHash string) error {
				if _, ok := registered[email]; ok {
					return storage.ErrUserExist
				}
				return nil
			})
		}
		suite.registerHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test name: %s", tt.name)
		if tt.wantSave {
			registered[tt.req.Email] = struct{}{}
		}
	}
}

func (suite *usrTestSuite) Test_Login() {
	tests := []struct{
		name string
		req httpmodels.LoginRequest
		registered bool
		wantCode int
		expectResponse bool
	}{
		{
			name: "happy_pass",
			req: httpmodels.LoginRequest{
				Email: "test@test.test",
				Password: "12345",
			},
			registered: true,
			wantCode: http.StatusOK,
			expectResponse: true,
		},
		{
			name: "user_not_exist",
			req: httpmodels.LoginRequest{
				Email: "test@test.test",
				Password: "12345",
			},
			registered: false,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		var jsonBody bytes.Buffer
		err := json.NewEncoder(&jsonBody).Encode(&tt.req)
		suite.Require().NoError(err, "test: %s, failed to encode request", tt.name)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/user/login", &jsonBody)
		passHash, err := bcrypt.GenerateFromPassword([]byte(tt.req.Password),bcrypt.DefaultCost)
		suite.Require().NoError(err, "test: %s, failed to hash password", tt.name)
		if tt.registered {
			suite.userRepoMock.
			On("SaveRefreshToken", mock.Anything, tt.req.Email, mock.AnythingOfType("string"), mock.AnythingOfType("int64")).
			Once().Return(nil)
		}
		suite.userRepoMock.On("GetUserByEmail", mock.Anything, tt.req.Email).Once().
		Return(func (ctx context.Context, email string) (models.User, error) {
			if tt.registered {
				return models.User{
					Email: tt.req.Email,
					PassHash: string(passHash),
				}, nil
			}
			return models.User{}, storage.ErrQuery
		})
		suite.loginHandler.ServeHTTP(w, r)
		suite.Require().Equal(tt.wantCode, w.Code, "test name: %s", tt.name)
		if tt.expectResponse {
			var resp httpmodels.LoginResponse
			err := json.NewDecoder(w.Body).Decode(&resp)
			suite.Require().NoError(err, "test name: %s", tt.name)
			claims, err := suite.jwtManager.ParseJwt(resp.TokenPair.AccessToken)
			suite.Require().NoError(err, "test name: %s", tt.name)
			suite.Equal(claims["email"], tt.req.Email, "test name: %s", tt.name)
			suite.InDelta(time.Now().Add(suite.cfg.TokenTTL).Unix(), claims["exp"].(float64), 1, "test name: %s", tt.name)
		}
	}
}

func (suite *usrTestSuite) Test_DoubleRefreshHappyPass() {
	ref, err := suite.jwtManager.CreateRefresh()
	suite.Require().NoError(err)
	hash, err := bcrypt.GenerateFromPassword([]byte(ref), bcrypt.DefaultCost)
	suite.Require().NoError(err)
	user := models.User{
		Email: "test@test.test",
		RefreshHash: string(hash),
		ExpiresAt: time.Now().Add(suite.cfg.RefreshTTL).Unix(),
	}
	token, err := suite.jwtManager.CreateJWT(user, suite.cfg.TokenTTL)
	suite.Require().NoError(err)
	var jsonBody bytes.Buffer
	req := httpmodels.RefreshRequest{
		TokenPair: models.TokenPair{
			AccessToken: token,
			RefreshToken: ref,
		},
	}
	err = json.NewEncoder(&jsonBody).Encode(&req)
	suite.Require().NoError(err, "failed to encode request")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/refresh", &jsonBody)
	suite.userRepoMock.On("GetUserByEmail", mock.Anything, mock.AnythingOfType("string")).Once().
	Return(func (ctx context.Context, email string) (models.User, error) {
		return user, nil
	})
	suite.userRepoMock.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64")).
	Once().Return(func(ctx context.Context, email string, refreshHash string, rttl int64) error {
		user.RefreshHash = refreshHash
		user.ExpiresAt = rttl
		return nil
	})
	suite.refreshHanlder.ServeHTTP(w,r)
	suite.Require().Equal(http.StatusOK, w.Code)
	var resp httpmodels.RefreshResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	suite.Require().NoError(err, "failed to decode response")
	newRef := resp.TokenPair.RefreshToken

	// request with old refresh token
	err = json.NewEncoder(&jsonBody).Encode(&req)
	suite.Require().NoError(err, "failed to encode request")
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/api/user/refresh", &jsonBody)
	suite.userRepoMock.On("GetUserByEmail", mock.Anything, mock.AnythingOfType("string")).Once().
	Return(func (ctx context.Context, email string) (models.User, error) {
		return user, nil
	})
	suite.refreshHanlder.ServeHTTP(w,r)
	suite.Require().Equal(http.StatusBadRequest, w.Code)

	// request with new refresh token
	req.TokenPair.RefreshToken = newRef
	err = json.NewEncoder(&jsonBody).Encode(&req)
	suite.Require().NoError(err, "failed to encode request")
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/api/user/refresh", &jsonBody)
	suite.userRepoMock.On("GetUserByEmail", mock.Anything, mock.AnythingOfType("string")).Once().
	Return(func (ctx context.Context, email string) (models.User, error) {
		return user, nil
	})
	suite.userRepoMock.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64")).
	Once().Return(nil)
	suite.refreshHanlder.ServeHTTP(w,r)
	suite.Require().Equal(http.StatusOK, w.Code)
}

func (suite *usrTestSuite) Test_AccessTokenExpired() {
	ref, err := suite.jwtManager.CreateRefresh()
	suite.Require().NoError(err)
	hash, err := bcrypt.GenerateFromPassword([]byte(ref), bcrypt.DefaultCost)
	suite.Require().NoError(err)
	user := models.User{
		Email: "test@test.test",
		RefreshHash: string(hash),
		ExpiresAt: time.Now().Add(suite.cfg.RefreshTTL).Unix(),
	}
	token, err := suite.jwtManager.CreateJWT(user, 0)
	suite.Require().NoError(err)
	var jsonBody bytes.Buffer
	req := httpmodels.RefreshRequest{
		TokenPair: models.TokenPair{
			AccessToken: token,
			RefreshToken: ref,
		},
	}
	err = json.NewEncoder(&jsonBody).Encode(&req)
	suite.Require().NoError(err, "failed to encode request")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/refresh", &jsonBody)
	suite.userRepoMock.On("GetUserByEmail", mock.Anything, mock.AnythingOfType("string")).Once().
	Return(func (ctx context.Context, email string) (models.User, error) {
		return user, nil
	})
	suite.userRepoMock.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("int64")).
	Once().Return(func(ctx context.Context, email string, refreshHash string, rttl int64) error {
		user.RefreshHash = refreshHash
		user.ExpiresAt = rttl
		return nil
	})
	suite.refreshHanlder.ServeHTTP(w,r)
	suite.Require().Equal(http.StatusOK, w.Code)
	var resp httpmodels.RefreshResponse
	err = json.NewDecoder(w.Body).Decode(&resp)
	suite.Require().NoError(err, "failed to decode response")
	suite.NotEmpty(resp.TokenPair.AccessToken)
	suite.NotEmpty(resp.TokenPair.RefreshToken)
}

func (suite *usrTestSuite) Test_RefreshTokenExpired() {
	ref, err := suite.jwtManager.CreateRefresh()
	suite.Require().NoError(err)
	hash, err := bcrypt.GenerateFromPassword([]byte(ref), bcrypt.DefaultCost)
	suite.Require().NoError(err)
	user := models.User{
		Email: "test@test.test",
		RefreshHash: string(hash),
		ExpiresAt: 0,
	}
	token, err := suite.jwtManager.CreateJWT(user, suite.cfg.TokenTTL)
	suite.Require().NoError(err)
	var jsonBody bytes.Buffer
	req := httpmodels.RefreshRequest{
		TokenPair: models.TokenPair{
			AccessToken: token,
			RefreshToken: ref,
		},
	}
	err = json.NewEncoder(&jsonBody).Encode(&req)
	suite.Require().NoError(err, "failed to encode request")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/user/refresh", &jsonBody)
	suite.userRepoMock.On("GetUserByEmail", mock.Anything, mock.AnythingOfType("string")).Once().
	Return(func (ctx context.Context, email string) (models.User, error) {
		return user, nil
	})
	suite.refreshHanlder.ServeHTTP(w,r)
	suite.Require().Equal(http.StatusBadRequest, w.Code)
}