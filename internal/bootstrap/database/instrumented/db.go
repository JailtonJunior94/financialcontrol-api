package instrumented

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

const dbSystemMSSQL = "mssql"

// Wrap returns a new *sql.DB that emits an OTel span for every BeginTx,
// QueryContext and ExecContext call. Attributes: db.system=mssql, db.name,
// db.operation.name, db.statement (sanitized via red). Emits a WARN log when
// duration exceeds threshold. Implemented via database/sql/driver wrapping.
func Wrap(
	db *sql.DB,
	dbName string,
	threshold time.Duration,
	red *redactor.Denylist,
	log *slog.Logger,
	tracer trace.Tracer,
) *sql.DB {
	return sql.OpenDB(&connector{
		source:    db,
		dbName:    dbName,
		threshold: threshold,
		red:       red,
		logger:    log,
		tracer:    tracer,
		now:       time.Now,
	})
}

// ------------------------------------------------------------------ connector

type connector struct {
	source    *sql.DB
	dbName    string
	threshold time.Duration
	red       *redactor.Denylist
	logger    *slog.Logger
	tracer    trace.Tracer
	now       func() time.Time
}

// Connect borrows a connection from the source pool and holds the underlying
// driver.Conn via sql.Conn.Raw for the lifetime of the returned tracedConn.
// The goroutine is unblocked when tracedConn.Close is called.
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	sqlConn, err := c.source.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("instrumented: borrow connection: %w", err)
	}

	// Buffered so the goroutine never blocks if the main goroutine exits early.
	connCh := make(chan driver.Conn, 1)
	releaseCh := make(chan struct{})

	go func() {
		_ = sqlConn.Raw(func(dc any) error {
			if driverConn, ok := dc.(driver.Conn); ok {
				connCh <- driverConn
			} else {
				connCh <- nil
			}
			<-releaseCh
			return nil
		})
	}()

	select {
	case inner := <-connCh:
		if inner == nil {
			close(releaseCh)
			_ = sqlConn.Close()
			return nil, fmt.Errorf("instrumented: underlying value does not implement driver.Conn")
		}
		return &tracedConn{
			inner:     inner,
			sqlConn:   sqlConn,
			releaseCh: releaseCh,
			dbName:    c.dbName,
			threshold: c.threshold,
			red:       c.red,
			logger:    c.logger,
			tracer:    c.tracer,
			now:       c.now,
		}, nil
	case <-ctx.Done():
		close(releaseCh)
		_ = sqlConn.Close()
		return nil, ctx.Err()
	}
}

func (c *connector) Driver() driver.Driver { return c.source.Driver() }

// ------------------------------------------------------------------ tracedConn

type tracedConn struct {
	inner     driver.Conn
	sqlConn   *sql.Conn
	releaseCh chan struct{}
	once      sync.Once

	dbName    string
	threshold time.Duration
	red       *redactor.Denylist
	logger    *slog.Logger
	tracer    trace.Tracer
	now       func() time.Time

	// transaction state — set by BeginTx, cleared by Commit/Rollback
	mu    sync.Mutex
	txCtx context.Context
}

func (t *tracedConn) Prepare(query string) (driver.Stmt, error) {
	return t.inner.Prepare(query)
}

func (t *tracedConn) Close() error {
	t.once.Do(func() { close(t.releaseCh) })
	return t.sqlConn.Close()
}

// Begin satisfies driver.Conn (legacy); delegates to BeginTx with empty options.
func (t *tracedConn) Begin() (driver.Tx, error) {
	return t.BeginTx(context.Background(), driver.TxOptions{})
}

// BeginTx starts a parent span for the transaction. Subsequent queries on the
// same connection use this span as their parent context.
func (t *tracedConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	start := t.now()
	stmt := "BEGIN"
	sanitized := t.red.RedactStatement(stmt)

	ctx, span := t.tracer.Start(ctx, "db.begin_tx", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", dbSystemMSSQL),
		attribute.String("db.name", t.dbName),
		attribute.String("db.operation.name", "BEGIN"),
		attribute.String("db.statement", sanitized),
	)

	var inner driver.Tx
	var err error
	if cbt, ok := t.inner.(driver.ConnBeginTx); ok {
		inner, err = cbt.BeginTx(ctx, opts)
	} else {
		inner, err = t.inner.Begin() //nolint:staticcheck
	}

	elapsed := t.now().Sub(start)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	t.maybeWarnSlow(ctx, elapsed, "BEGIN", sanitized, span)

	// Store tx context so child queries inherit this span as parent.
	t.mu.Lock()
	t.txCtx = ctx
	t.mu.Unlock()

	return &tracedTx{inner: inner, conn: t, span: span}, nil
}

// QueryContext intercepts SELECT/other read queries.
func (t *tracedConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	ctx = t.activeTxCtx(ctx)
	op := extractOperation(query)
	sanitized := t.red.RedactStatement(query)

	start := t.now()
	ctx, span := t.tracer.Start(ctx, "db.query", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", dbSystemMSSQL),
		attribute.String("db.name", t.dbName),
		attribute.String("db.operation.name", op),
		attribute.String("db.statement", sanitized),
	)

	var rows driver.Rows
	var err error
	if qc, ok := t.inner.(driver.QueryerContext); ok {
		rows, err = qc.QueryContext(ctx, query, args)
	} else {
		rows, err = queryViaStmt(t.inner, ctx, query, args)
	}

	elapsed := t.now().Sub(start)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.End()
	t.maybeWarnSlow(ctx, elapsed, op, sanitized, span)
	return rows, nil
}

// ExecContext intercepts INSERT/UPDATE/DELETE and other write queries.
func (t *tracedConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	ctx = t.activeTxCtx(ctx)
	op := extractOperation(query)
	sanitized := t.red.RedactStatement(query)

	start := t.now()
	ctx, span := t.tracer.Start(ctx, "db.exec", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", dbSystemMSSQL),
		attribute.String("db.name", t.dbName),
		attribute.String("db.operation.name", op),
		attribute.String("db.statement", sanitized),
	)

	var result driver.Result
	var err error
	if ec, ok := t.inner.(driver.ExecerContext); ok {
		result, err = ec.ExecContext(ctx, query, args)
	} else {
		result, err = execViaStmt(t.inner, ctx, query, args)
	}

	elapsed := t.now().Sub(start)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	span.End()
	t.maybeWarnSlow(ctx, elapsed, op, sanitized, span)
	return result, nil
}

// activeTxCtx returns txCtx when inside a transaction, otherwise returns ctx.
func (t *tracedConn) activeTxCtx(ctx context.Context) context.Context {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.txCtx != nil {
		return t.txCtx
	}
	return ctx
}

// maybeWarnSlow emits a WARN log when elapsed > threshold.
func (t *tracedConn) maybeWarnSlow(ctx context.Context, elapsed time.Duration, op, sanitized string, span trace.Span) {
	if elapsed <= t.threshold {
		return
	}
	sc := span.SpanContext()
	t.logger.WarnContext(ctx, "slow_query",
		slog.String("event", "slow_query"),
		slog.Int64("query_duration_ms", elapsed.Milliseconds()),
		slog.String("db.operation.name", op),
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
		slog.String("db.statement", sanitized),
	)
}

// ------------------------------------------------------------------ tracedTx

type tracedTx struct {
	inner driver.Tx
	conn  *tracedConn
	span  trace.Span
}

func (t *tracedTx) Commit() error {
	err := t.inner.Commit()
	t.conn.mu.Lock()
	t.conn.txCtx = nil
	t.conn.mu.Unlock()
	t.span.End()
	return err
}

func (t *tracedTx) Rollback() error {
	err := t.inner.Rollback()
	t.conn.mu.Lock()
	t.conn.txCtx = nil
	t.conn.mu.Unlock()
	t.span.End()
	return err
}

// ------------------------------------------------------------------ helpers

// extractOperation returns the SQL verb (SELECT, INSERT, UPDATE, DELETE, …)
// from the first token of the query. Returns "UNKNOWN" when the query is empty.
func extractOperation(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return "UNKNOWN"
	}
	idx := strings.IndexAny(q, " \t\n\r(")
	if idx < 0 {
		return strings.ToUpper(q)
	}
	return strings.ToUpper(q[:idx])
}

// queryViaStmt falls back to Prepare+Query when the driver does not implement
// driver.QueryerContext.
func queryViaStmt(conn driver.Conn, ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = stmt.Close() }()

	dargs, err := namedToValues(args)
	if err != nil {
		return nil, err
	}

	if qs, ok := stmt.(driver.StmtQueryContext); ok {
		return qs.QueryContext(ctx, args)
	}
	return stmt.Query(dargs) //nolint:staticcheck
}

// execViaStmt falls back to Prepare+Exec when the driver does not implement
// driver.ExecerContext.
func execViaStmt(conn driver.Conn, ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	stmt, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = stmt.Close() }()

	dargs, err := namedToValues(args)
	if err != nil {
		return nil, err
	}

	if es, ok := stmt.(driver.StmtExecContext); ok {
		return es.ExecContext(ctx, args)
	}
	return stmt.Exec(dargs) //nolint:staticcheck
}

// namedToValues converts []driver.NamedValue to []driver.Value for legacy interfaces.
func namedToValues(args []driver.NamedValue) ([]driver.Value, error) {
	vals := make([]driver.Value, len(args))
	for i, nv := range args {
		vals[i] = nv.Value
	}
	return vals, nil
}
