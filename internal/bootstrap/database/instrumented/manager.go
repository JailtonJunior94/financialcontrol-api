package instrumented

import (
	"context"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	devkitmanager "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
)

// WrapManager decorates a manager.Manager so every DBTX/Tx operation emits OTel spans
// with sanitized db.statement and slow-query WARN logs.
func WrapManager(
	inner devkitmanager.Manager,
	obs observability.Observability,
	dbName string,
	threshold time.Duration,
	dl *redactor.Denylist,
) devkitmanager.Manager {
	if inner == nil || obs == nil || dl == nil {
		return inner
	}
	return &instrumentedManager{
		inner:     inner,
		obs:       obs,
		driver:    string(inner.Driver()),
		dbName:    dbName,
		threshold: threshold,
		denylist:  dl,
		now:       time.Now,
	}
}

type instrumentedManager struct {
	inner     devkitmanager.Manager
	obs       observability.Observability
	driver    string
	dbName    string
	threshold time.Duration
	denylist  *redactor.Denylist
	now       func() time.Time
}

func (m *instrumentedManager) Driver() devkitdb.Driver { return m.inner.Driver() }

func (m *instrumentedManager) DBTX(ctx context.Context) devkitdb.DBTX {
	if tx, ok := devkitdb.FromContext(ctx); ok {
		return tx
	}
	return &instrumentedDBTX{
		base:      m.inner.DBTX(ctx),
		obs:       m.obs,
		driver:    m.driver,
		dbName:    m.dbName,
		threshold: m.threshold,
		denylist:  m.denylist,
		now:       m.now,
	}
}

func (m *instrumentedManager) BeginTx(ctx context.Context, opts devkitdb.TxOptions) (devkitdb.Tx, error) {
	txCtx, span := m.startSpan(ctx, "db.begin_tx", "BEGIN")
	tx, err := m.inner.BeginTx(txCtx, opts)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(observability.StatusCodeError, err.Error())
		span.End()
		return nil, err
	}

	span.SetStatus(observability.StatusCodeOK, "ok")
	return &instrumentedTx{
		base: tx,
		dbtx: instrumentedDBTX{
			base:      tx,
			obs:       m.obs,
			driver:    m.driver,
			dbName:    m.dbName,
			threshold: m.threshold,
			denylist:  m.denylist,
			now:       m.now,
			txCtx:     txCtx,
		},
		span: span,
	}, nil
}

func (m *instrumentedManager) Ping(ctx context.Context) error     { return m.inner.Ping(ctx) }
func (m *instrumentedManager) Shutdown(ctx context.Context) error { return m.inner.Shutdown(ctx) }

func (m *instrumentedManager) startSpan(ctx context.Context, name, query string) (context.Context, observability.Span) {
	sanitized := m.denylist.RedactStatement(query)
	op := extractOperation(query)
	return m.obs.Tracer().Start(ctx, name,
		observability.WithSpanKind(observability.SpanKindClient),
		observability.WithAttributes(
			observability.String("db.system", m.driver),
			observability.String("db.name", m.dbName),
			observability.String("db.operation.name", op),
			observability.String("db.statement", sanitized),
		),
	)
}

type instrumentedDBTX struct {
	base      devkitdb.DBTX
	obs       observability.Observability
	driver    string
	dbName    string
	threshold time.Duration
	denylist  *redactor.Denylist
	now       func() time.Time
	txCtx     context.Context
}

func (d *instrumentedDBTX) ExecContext(ctx context.Context, query string, args ...any) (devkitdb.Result, error) {
	ctx, span, sanitized, started := d.start(ctx, "db.exec", query)
	result, err := d.base.ExecContext(ctx, query, args...)
	d.finish(ctx, span, extractOperation(query), sanitized, started, err)
	return result, err
}

func (d *instrumentedDBTX) QueryContext(ctx context.Context, query string, args ...any) (devkitdb.Rows, error) {
	ctx, span, sanitized, started := d.start(ctx, "db.query", query)
	rows, err := d.base.QueryContext(ctx, query, args...)
	if err != nil {
		d.finish(ctx, span, extractOperation(query), sanitized, started, err)
		return nil, err
	}

	return &instrumentedRows{
		base:      rows,
		ctx:       ctx,
		span:      span,
		op:        extractOperation(query),
		statement: sanitized,
		started:   started,
		parent:    d,
	}, nil
}

func (d *instrumentedDBTX) QueryRowContext(ctx context.Context, query string, args ...any) devkitdb.Row {
	ctx, span, sanitized, started := d.start(ctx, "db.query_row", query)
	row := d.base.QueryRowContext(ctx, query, args...)
	return &instrumentedRow{
		base:      row,
		ctx:       ctx,
		span:      span,
		op:        extractOperation(query),
		statement: sanitized,
		started:   started,
		parent:    d,
	}
}

func (d *instrumentedDBTX) start(ctx context.Context, spanName, query string) (context.Context, observability.Span, string, time.Time) {
	parent := ctx
	if d.txCtx != nil {
		parent = d.txCtx
	}

	sanitized := d.denylist.RedactStatement(query)
	op := extractOperation(query)
	ctx, span := d.obs.Tracer().Start(parent, spanName,
		observability.WithSpanKind(observability.SpanKindClient),
		observability.WithAttributes(
			observability.String("db.system", d.driver),
			observability.String("db.name", d.dbName),
			observability.String("db.operation.name", op),
			observability.String("db.statement", sanitized),
		),
	)
	return ctx, span, sanitized, d.now()
}

func (d *instrumentedDBTX) finish(ctx context.Context, span observability.Span, op, statement string, started time.Time, err error) {
	elapsed := d.now().Sub(started)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(observability.StatusCodeError, err.Error())
	} else {
		span.SetStatus(observability.StatusCodeOK, "ok")
	}
	span.End()
	d.warnSlow(ctx, span, elapsed, op, statement)
}

func (d *instrumentedDBTX) warnSlow(ctx context.Context, span observability.Span, elapsed time.Duration, op, statement string) {
	if elapsed <= d.threshold {
		return
	}

	d.obs.Logger().Warn(ctx, "slow_query",
		observability.String("event", "slow_query"),
		observability.Int64("query_duration_ms", elapsed.Milliseconds()),
		observability.String("db.operation.name", op),
		observability.String("trace_id", span.TraceID()),
		observability.String("span_id", span.SpanID()),
		observability.String("db.statement", statement),
	)
}

type instrumentedTx struct {
	base devkitdb.Tx
	dbtx instrumentedDBTX
	span observability.Span
}

func (t *instrumentedTx) ExecContext(ctx context.Context, query string, args ...any) (devkitdb.Result, error) {
	return t.dbtx.ExecContext(ctx, query, args...)
}

func (t *instrumentedTx) QueryContext(ctx context.Context, query string, args ...any) (devkitdb.Rows, error) {
	return t.dbtx.QueryContext(ctx, query, args...)
}

func (t *instrumentedTx) QueryRowContext(ctx context.Context, query string, args ...any) devkitdb.Row {
	return t.dbtx.QueryRowContext(ctx, query, args...)
}

func (t *instrumentedTx) Commit(ctx context.Context) error {
	err := t.base.Commit(ctx)
	if err != nil {
		t.span.RecordError(err)
		t.span.SetStatus(observability.StatusCodeError, err.Error())
	} else {
		t.span.SetStatus(observability.StatusCodeOK, "ok")
	}
	t.span.End()
	return err
}

func (t *instrumentedTx) Rollback(ctx context.Context) error {
	err := t.base.Rollback(ctx)
	if err != nil {
		t.span.RecordError(err)
		t.span.SetStatus(observability.StatusCodeError, err.Error())
	} else {
		t.span.SetStatus(observability.StatusCodeOK, "ok")
	}
	t.span.End()
	return err
}

type instrumentedRows struct {
	base      devkitdb.Rows
	ctx       context.Context
	span      observability.Span
	op        string
	statement string
	started   time.Time
	parent    *instrumentedDBTX
	done      bool
}

func (r *instrumentedRows) Next() bool {
	next := r.base.Next()
	if !next {
		r.complete(r.base.Err())
	}
	return next
}

func (r *instrumentedRows) Scan(dest ...any) error { return r.base.Scan(dest...) }

func (r *instrumentedRows) Close() error {
	err := r.base.Close()
	r.complete(err)
	return err
}

func (r *instrumentedRows) Err() error {
	err := r.base.Err()
	if err != nil {
		r.complete(err)
	}
	return err
}

func (r *instrumentedRows) complete(err error) {
	if r.done {
		return
	}
	r.done = true
	r.parent.finish(r.ctx, r.span, r.op, r.statement, r.started, err)
}

type instrumentedRow struct {
	base      devkitdb.Row
	ctx       context.Context
	span      observability.Span
	op        string
	statement string
	started   time.Time
	parent    *instrumentedDBTX
	done      bool
}

func (r *instrumentedRow) Scan(dest ...any) error {
	err := r.base.Scan(dest...)
	if !r.done {
		r.done = true
		r.parent.finish(r.ctx, r.span, r.op, r.statement, r.started, err)
	}
	return err
}
