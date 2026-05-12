package hooks

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
)

// stubIssuer issues a deterministic fake token for tests.
type stubIssuer struct{ err error }

func (s *stubIssuer) Issue(_ context.Context, id pkgjwt.Identity) (string, time.Time, error) {
	if s.err != nil {
		return "", time.Time{}, s.err
	}
	return "fake-token-for-" + id.UserID, time.Now().Add(time.Hour), nil
}

// stubHTTP is a minimal testableHTTP that returns canned responses.
type stubHTTP struct {
	handler func(path string) (int, []byte)
}

func (s *stubHTTP) Test(req *http.Request, _ ...int) (*http.Response, error) {
	code, body := s.handler(req.URL.Path)
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

// errHTTP is a testableHTTP that always returns a transport error.
type errHTTP struct{ err error }

func (e *errHTTP) Test(_ *http.Request, _ ...int) (*http.Response, error) {
	return nil, e.err
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// buildSmoke constructs a financeSmoke with the given users and a stubHTTP handler.
func buildSmoke(
	users []string,
	handler func(path string) (int, []byte),
) *financeSmoke {
	var auditWritten bool
	_ = auditWritten
	return &financeSmoke{
		http:   &stubHTTP{handler: handler},
		issuer: &stubIssuer{},
		logger: newDiscardLogger(),
		queryUsers: func(_ context.Context) ([]string, error) { return users, nil },
		writeAudit: func(_ context.Context) error { return nil },
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Table-driven tests
// ──────────────────────────────────────────────────────────────────────────────

func TestFinanceSmoke_GoldenPath_NoUsers(t *testing.T) {
	t.Parallel()
	s := buildSmoke(nil, func(path string) (int, []byte) {
		t.Errorf("unexpected request to %s — no users expected", path)
		return 200, nil
	})
	require.NoError(t, s.run(context.Background()))
}

func TestFinanceSmoke_GoldenPath_OneUser_ZeroInvoices(t *testing.T) {
	t.Parallel()

	emptyList := []byte(`{"items":[],"total":0,"page":1,"page_size":20}`)
	called := map[string]int{}

	s := buildSmoke([]string{"user-abc"}, func(path string) (int, []byte) {
		called[path]++
		return 200, emptyList
	})

	require.NoError(t, s.run(context.Background()))
	assert.Equal(t, 1, called["/api/v1/finance/transactions"], "list transactions called once")
	assert.Equal(t, 1, called["/api/v1/finance/summary"], "summary called once")
	assert.Equal(t, 1, called["/api/v1/finance/invoices"], "list invoices called once")
	// total == 0 → no invoice detail call
	for k := range called {
		if k != "/api/v1/finance/transactions" &&
			k != "/api/v1/finance/summary" &&
			k != "/api/v1/finance/invoices" {
			t.Errorf("unexpected path called: %s", k)
		}
	}
}

func TestFinanceSmoke_GoldenPath_OneUser_WithInvoice(t *testing.T) {
	t.Parallel()

	invoiceID := "inv-id-123"
	invoiceList := []byte(`{"items":[{"id":"` + invoiceID + `"}],"total":1,"page":1,"page_size":20}`)
	emptyList := []byte(`{"items":[],"total":0,"page":1,"page_size":20}`)
	called := map[string]int{}

	s := buildSmoke([]string{"user-abc"}, func(path string) (int, []byte) {
		called[path]++
		if path == "/api/v1/finance/invoices" {
			return 200, invoiceList
		}
		return 200, emptyList
	})

	require.NoError(t, s.run(context.Background()))
	assert.Equal(t, 1, called["/api/v1/finance/invoices/"+invoiceID],
		"invoice detail called for the returned invoice ID")
}

func TestFinanceSmoke_Failure_ListTransactions_Non2xx(t *testing.T) {
	t.Parallel()

	s := buildSmoke([]string{"user-abc"}, func(path string) (int, []byte) {
		if path == "/api/v1/finance/transactions" {
			return 500, []byte(`{"error":"internal"}`)
		}
		return 200, []byte(`{"items":[],"total":0,"page":1,"page_size":20}`)
	})

	err := s.run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list transactions")
}

func TestFinanceSmoke_Failure_Summary_Non2xx(t *testing.T) {
	t.Parallel()

	s := buildSmoke([]string{"user-abc"}, func(path string) (int, []byte) {
		if path == "/api/v1/finance/summary" {
			return 403, []byte(`{"error":"forbidden"}`)
		}
		return 200, []byte(`{"items":[],"total":0,"page":1,"page_size":20}`)
	})

	err := s.run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "monthly summary")
}

func TestFinanceSmoke_Failure_TransportError(t *testing.T) {
	t.Parallel()

	s := &financeSmoke{
		http:       &errHTTP{err: errors.New("connection refused")},
		issuer:     &stubIssuer{},
		logger:     newDiscardLogger(),
		queryUsers: func(_ context.Context) ([]string, error) { return []string{"u1"}, nil },
		writeAudit: func(_ context.Context) error { return nil },
	}

	err := s.run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestFinanceSmoke_Failure_QueryUsers(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db unreachable")
	s := &financeSmoke{
		http:       &stubHTTP{handler: func(_ string) (int, []byte) { return 200, nil }},
		issuer:     &stubIssuer{},
		logger:     newDiscardLogger(),
		queryUsers: func(_ context.Context) ([]string, error) { return nil, dbErr },
		writeAudit: func(_ context.Context) error { return nil },
	}

	err := s.run(context.Background())
	require.ErrorIs(t, err, dbErr)
}

func TestFinanceSmoke_Failure_WriteAudit(t *testing.T) {
	t.Parallel()

	auditErr := errors.New("write failed")
	s := &financeSmoke{
		http:       &stubHTTP{handler: func(_ string) (int, []byte) { return 200, []byte(`{"items":[],"total":0}`) }},
		issuer:     &stubIssuer{},
		logger:     newDiscardLogger(),
		queryUsers: func(_ context.Context) ([]string, error) { return []string{"u1"}, nil },
		writeAudit: func(_ context.Context) error { return auditErr },
	}

	err := s.run(context.Background())
	require.ErrorIs(t, err, auditErr)
}

func TestFinanceSmoke_Failure_IssuerError(t *testing.T) {
	t.Parallel()

	issuerErr := errors.New("signing error")
	s := &financeSmoke{
		http:       &stubHTTP{handler: func(_ string) (int, []byte) { return 200, nil }},
		issuer:     &stubIssuer{err: issuerErr},
		logger:     newDiscardLogger(),
		queryUsers: func(_ context.Context) ([]string, error) { return []string{"u1"}, nil },
		writeAudit: func(_ context.Context) error { return nil },
	}

	err := s.run(context.Background())
	require.ErrorIs(t, err, issuerErr)
}
