package pgdb

import (
	dberrors "carshop/internal/database/errors"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxUUID "github.com/vgarvardt/pgx-google-uuid/v5"
	"go.uber.org/zap"
)

type Config struct {
	Hosts    string
	Port     uint16
	Username string
	Password string
	Database string
}

type client struct {
	config Config
	pool   *ConnPool
}

func NewClient(cfg Config, logger *zap.Logger) (*client, error) {
	connString := fmt.Sprintf(
		"port=%d dbname=%s user=%s password=%s pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s",
		cfg.Port,
		cfg.Database,
		cfg.Username,
		cfg.Password,
		DefaultMaxConns,
		DefaultMaxConnLifetime,
		DefaultMaxConnIdleTime,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	poolConfig.AfterConnect = func(ctx context.Context, pgxConn *pgx.Conn) error {
		pgxUUID.Register(pgxConn.TypeMap())
		return nil
	}

	poolConfig.ConnConfig.TLSConfig = nil

	ctx := context.Background()

	pool, err := NewConnPool(ctx, poolConfig, strings.Split(cfg.Hosts, ","), logger)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create connection pool: %w", dberrors.ErrInternal, err)
	}

	conn, err := pool.PeekWriteConn()
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("%w: failed to get connection from pool: %w", dberrors.ErrInternal, err)
	}
	defer conn.Release()

	if err := conn.Conn.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &client{
		config: cfg,
		pool:   pool,
	}, nil
}
