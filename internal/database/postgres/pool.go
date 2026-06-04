package pgdb

import (
	dberrors "carshop/internal/database/errors"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Driver interface {
	// QueryRow is a convenience wrapper over Query. Any error that occurs while
	// querying is deferred until calling Scan on the returned Row. That Row will
	// error with ErrNoRows if no rows are returned.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	// Query sends a query to the server and returns a Rows to read the results. Only errors encountered sending the query
	// and initializing Rows will be returned. Err() on the returned Rows must be checked after the Rows is closed to
	// determine if the query executed successfully.
	//
	// The returned Rows must be closed before the connection can be used again. It is safe to attempt to read from the
	// returned Rows even if an error is returned. The error will be the available in rows.Err() after rows are closed. It
	// is allowed to ignore the error returned from Query and handle it in Rows.
	//
	// It is possible for a call of FieldDescriptions on the returned Rows to return nil even if the Query call did not
	// return an error.
	//
	// It is possible for a query to return one or more rows before encountering an error. In most cases the rows should be
	// collected before processing rather than processed while receiving each row. This avoids the possibility of the
	// application processing rows from a query that the server rejected. The CollectRows function is useful here.
	//
	// An implementor of QueryRewriter may be passed as the first element of args. It can rewrite the sql and change or
	// replace args. For example, NamedArgs is QueryRewriter that implements named arguments.
	//
	// For extra control over how the query is executed, the types QueryExecMode, QueryResultFormats, and
	// QueryResultFormatsByOID may be used as the first args to control exactly how the query is executed. This is rarely
	// needed. See the documentation for those types for details.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// Exec executes sql. sql can be either a prepared statement name or an SQL string. arguments should be referenced
	// positionally from the sql string as $1, $2, etc.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Conn struct {
	*pgxpool.Conn
	Pool *ConnPool
}

func (c *Conn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.Conn.QueryRow(ctx, sql, args...)
}

func (c *Conn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := c.Conn.Query(ctx, sql, args...)
	if err != nil {
		c.checkROError(err)
	}

	return rows, err
}

func (c *Conn) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ctag, err := c.Conn.Exec(ctx, sql, args...)
	if err != nil {
		c.checkROError(err)
	}

	return ctag, err
}

func (c *Conn) checkROError(err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == PGErrROTransaction {
			_ = c.Pool.NotifyRO()
		}
	}
}

type ConnPool struct {
	writePool atomic.Pointer[pgxpool.Pool]
	readPools []*pgxpool.Pool
	readIdx   atomic.Uint32

	config *pgxpool.Config
	hosts  []string
	ctx    context.Context
	cancel context.CancelFunc

	mu             sync.Mutex
	initInProgress atomic.Bool

	logger *zap.Logger
}

func NewConnPool(ctx context.Context, connConfig *pgxpool.Config, hosts []string, logger *zap.Logger) (*ConnPool, error) {
	ctx, cancel := context.WithCancel(ctx)

	pool := &ConnPool{
		config: connConfig,
		hosts:  hosts,
		ctx:    ctx,
		cancel: cancel,

		logger: logger,
	}

	if err := pool.initPools(); err != nil {
		return nil, fmt.Errorf("error initializing pools: %w", err)
	}

	return pool, nil
}

func (p *ConnPool) initPools() error {
	p.logger.Info("initializing pools...")

	var (
		writePool *pgxpool.Pool
		readPools []*pgxpool.Pool
	)

	for _, host := range p.hosts {
		cfg := p.config.Copy()
		cfg.ConnConfig.Host = host

		pool, err := pgxpool.NewWithConfig(p.ctx, cfg)
		if err != nil {
			p.logger.With(
				zap.String("host", host),
				zap.String("error", err.Error()),
			).Error("failed to connect")

			continue
		}

		isPrimary, err := isPrimary(p.ctx, pool)
		if err != nil {
			pool.Close()
			p.logger.With(
				zap.String("host", host),
				zap.String("error", err.Error()),
			).Error("failed to check if primary")

			continue
		}

		if isPrimary && writePool == nil {
			writePool = pool
		} else {
			readPools = append(readPools, pool)
		}
	}

	if writePool == nil {
		return fmt.Errorf("no primary host found")
	}

	p.writePool.Store(writePool)

	p.mu.Lock()
	p.readPools = readPools
	p.mu.Unlock()

	p.logger.Info("successfully initialized connection pools")

	return nil
}

func (p *ConnPool) reinitWritePool() error {
	p.logger.Info("reinitializing read-write pool...")

	var writePool *pgxpool.Pool

	for _, host := range p.hosts {
		cfg := p.config.Copy()
		cfg.ConnConfig.Host = host

		pool, err := pgxpool.NewWithConfig(p.ctx, cfg)
		if err != nil {
			p.logger.With(
				zap.String("host", host),
				zap.String("error", err.Error()),
			).Error("failed to connect")

			continue
		}

		isPrimary, err := isPrimary(p.ctx, pool)
		if err != nil {
			pool.Close()
			p.logger.With(
				zap.String("host", host),
				zap.String("error", err.Error()),
			).Error("failed to check if primary")

			continue
		}

		if isPrimary {
			writePool = pool
			break
		}
	}

	if writePool == nil {
		return fmt.Errorf("no primary host found")
	}

	p.writePool.Store(writePool)

	return nil
}

func (p *ConnPool) NotifyRO() error {
	if !p.initInProgress.CompareAndSwap(false, true) {
		return nil
	}

	defer p.initInProgress.Store(false)

	if writePool := p.writePool.Load(); writePool != nil {
		writePool.Close()
	}

	if err := p.reinitWritePool(); err != nil {
		return err
	}

	return nil
}

func isPrimary(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var isPrimary bool

	err := pool.QueryRow(ctx, "SELECT NOT pg_is_in_recovery()").Scan(&isPrimary)

	return isPrimary, err
}

func (p *ConnPool) PeekWriteConn() (*Conn, error) {
	if p.initInProgress.Load() {
		return nil, dberrors.ErrUnavailable
	}

	if writePool := p.writePool.Load(); writePool != nil {
		conn, err := writePool.Acquire(p.ctx)
		if err != nil {
			return nil, err
		}

		return &Conn{Conn: conn, Pool: p}, nil
	}

	return nil, fmt.Errorf("no write connection available")
}

func (p *ConnPool) PeekReadConn() (*Conn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.readPools) == 0 {
		if writePool := p.writePool.Load(); writePool != nil {
			conn, err := writePool.Acquire(p.ctx)
			if err != nil {
				return nil, err
			}

			return &Conn{Conn: conn, Pool: p}, nil
		}

		return nil, fmt.Errorf("no read connection available")
	}

	idx := int(p.readIdx.Add(1)) % len(p.readPools)

	conn, err := p.readPools[idx].Acquire(p.ctx)
	if err != nil {
		return nil, err
	}

	return &Conn{Conn: conn, Pool: p}, nil
}

func (p *ConnPool) Close() {
	p.cancel()
	p.closePools()
}

func (p *ConnPool) closePools() {
	if writePool := p.writePool.Load(); writePool != nil {
		writePool.Close()
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.readPools {
		pool.Close()
	}

	p.readPools = nil
}
