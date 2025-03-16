package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/EwvwGeN/cataloger/internal/domain/models"
	"github.com/jackc/pgconn"
)

func (pp *postgresProvider) SaveUser(ctx context.Context, email string, passHash string) (error) {
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf(`INSERT INTO "%s" (email, pass_hash)
VALUES($1,$2);`, pp.cfg.UserTable), email, passHash)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrUserExist
		}
	}
	return ErrQuery
}

func (pp *postgresProvider) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	row := pp.dbConn.QueryRow(ctx, fmt.Sprintf(`
SELECT "email", "pass_hash", "refresh_hash", "expires_at"
FROM "%s"
WHERE "email"=$1;`,
	pp.cfg.UserTable),
	email)
	// better to use storage layer model and then convert it intro service layer
	// or user pointer in service model
	var (
		user models.User
		refHash *string
		expiresAt *int64
	)
	err := row.Scan(&user.Email, &user.PassHash, &refHash, &expiresAt)
	if err != nil {
		return models.User{}, ErrQuery
	}
	if refHash != nil {
		user.RefreshHash = *refHash
	}
	if expiresAt != nil {
		user.ExpiresAt = *expiresAt
	}
	return user, nil
}

func (pp *postgresProvider) SaveRefreshToken(ctx context.Context, email string, refreshToken string, rttl int64) error {
	_, err := pp.dbConn.Exec(ctx, fmt.Sprintf(`
UPDATE "%s" SET
"refresh_hash" = $1, "expires_at" = $2
WHERE "email" = $3`,
	pp.cfg.UserTable),
	refreshToken,
	rttl,
	email,
	)
	if err != nil {
		return ErrQuery
	}
	return nil
}