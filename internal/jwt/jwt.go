package jwt

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/EwvwGeN/cataloger/internal/domain/models"
	"github.com/golang-jwt/jwt"
)

type jwtManager struct {
	secretKey string
	randSouce rand.Source
}

func NewJwtManager(secretKey string) *jwtManager {
	return &jwtManager{
		secretKey: secretKey,
		randSouce: rand.NewSource(time.Now().Unix()),
	}
}

func (jm *jwtManager) CreateJWT(user models.User, ttl time.Duration) (token string, err error) {
	if user.Email == "" {
		return "", ErrEmptyValue
	}
	tokenObject := jwt.New(jwt.SigningMethodHS512)
	claims := tokenObject.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(ttl).Unix()
	token, err = tokenObject.SignedString([]byte(jm.secretKey))
	if err != nil {
		return "", err
	}
	return
}

func (jm *jwtManager) CreateRefresh() (refresh string, err error) {
	buffer := make([]byte, 32)
	gen := rand.New(jm.randSouce)
	if _, err = gen.Read(buffer); err != nil {
		return "", ErrRefreshGenerate
	}
	return fmt.Sprintf("%x", buffer), nil
}

func (jm *jwtManager) MustParseJwt(token string) (map[string]interface{}, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(jm.secretKey), nil
	})
	var validErr *jwt.ValidationError
	if errors.As(err, &validErr) {
		if !(validErr.Errors == jwt.ValidationErrorExpired) {
			return nil, err
		}
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrParseClaims
	}
	return claims, nil
}

func (jm *jwtManager) ParseJwt(token string) (map[string]interface{}, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(jm.secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrParseClaims
	}
	return claims, nil
}