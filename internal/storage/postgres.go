package storage

import (
	"context"
	"fmt"

	"github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	"github.com/jackc/pgx/v4"
)

type postgresProvider struct {
	cfg config.PostgresConfig
	dbConn *pgx.Conn
}

func NewPostgresProvider(ctx context.Context, cfg config.PostgresConfig) (*postgresProvider, error) {
	connString := fmt.Sprintf("%s://%s:%s@%s:%s/%s",
		cfg.ConectionFormat,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql: %w", err)
	}
	return &postgresProvider{
		cfg: cfg,
		dbConn: conn,
	}, nil
}