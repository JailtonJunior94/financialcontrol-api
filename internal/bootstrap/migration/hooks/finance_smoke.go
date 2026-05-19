package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	bootstrapcontainer "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/container"
	bootstraphttp "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/http"
	bootstrapmigration "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
)

func init() {
	bootstrapmigration.RegisterHook("finance", FinanceSmokeHook)
}

// closeRuntime releases the runtime resources owned by the smoke hook: the
// observability provider (OTel batch processors, log exporters) and the database
// manager. Both shutdowns share a bounded 5s budget. Errors are swallowed — the
// hook return value already carries the smoke outcome and these shutdowns happen
// during defer, so logging here would mask the smoke result.
func closeRuntime(c *bootstrapcontainer.Container) {
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = c.Observability.Shutdown(shutCtx)
	_ = c.DBManager.Shutdown(shutCtx)
}

// testableHTTP abstracts Fiber's app.Test so the hook can be tested without a real server.
type testableHTTP interface {
	Test(req *http.Request, msTimeout ...int) (*http.Response, error)
}

// financeSmoke holds injected dependencies so the smoke logic is testable.
type financeSmoke struct {
	http       testableHTTP
	issuer     pkgjwt.Issuer
	logger     *slog.Logger
	queryUsers func(ctx context.Context) ([]string, error)
	writeAudit func(ctx context.Context) error
}

type closableRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

// smokeInvoiceList is the minimal shape of GET /finance/invoices for parsing total + first ID.
type smokeInvoiceList struct {
	Total int64              `json:"total"`
	Items []smokeInvoiceItem `json:"items"`
}

type smokeInvoiceItem struct {
	ID string `json:"id"`
}

// FinanceSmokeHook is the production entry-point registered in init().
// It builds the full runtime, runs smoke requests for every user that has at least
// one finance.Transaction, and writes Status='smoke_ok' to dbo.FinanceMigrationAudit
// on success.  Any non-2xx response or infrastructure error returns a non-nil error.
func FinanceSmokeHook(ctx context.Context, logger *slog.Logger) error {
	c, err := bootstrapcontainer.BuildRuntime(ctx)
	if err != nil {
		return fmt.Errorf("build runtime: %w", err)
	}
	defer closeRuntime(c)

	jwtCfg := pkgjwt.Config{
		Secret:    []byte(config.JwtSecret),
		AccessTTL: time.Hour * time.Duration(config.ExpirationAt),
	}
	issuer, err := pkgjwt.NewIssuer(jwtCfg)
	if err != nil {
		return fmt.Errorf("build jwt issuer: %w", err)
	}

	srv, err := bootstraphttp.NewServer(c)
	if err != nil {
		return fmt.Errorf("build http server: %w", err)
	}

	db := c.DBManager.DBTX(ctx)

	s := &financeSmoke{
		http:   srv.App(),
		issuer: issuer,
		logger: logger,
		queryUsers: func(qCtx context.Context) ([]string, error) {
			rows, err := db.QueryContext(qCtx,
				`SELECT DISTINCT CAST([UserId] AS VARCHAR(36))
				 FROM dbo.FinanceTransactions
				 WHERE [DeletedAt] IS NULL`)
			if err != nil {
				return nil, err
			}
			return collectUserIDs(rows)
		},
		writeAudit: func(aCtx context.Context) error {
			_, err := db.ExecContext(aCtx, `
				INSERT INTO dbo.FinanceMigrationAudit
				    ([Id],[StartedAt],[FinishedAt],[SourceTable],[RowsRead],[RowsWritten],[Status])
				VALUES (NEWID(),GETUTCDATE(),GETUTCDATE(),'*',0,0,'smoke_ok')`)
			return err
		},
	}

	return s.run(ctx)
}

// run executes the smoke sequence: list of users → HTTP probes → audit write.
func (s *financeSmoke) run(ctx context.Context) error {
	users, err := s.queryUsers(ctx)
	if err != nil {
		return fmt.Errorf("query users with transactions: %w", err)
	}

	s.logger.Info("finance smoke starting", slog.Int("user_count", len(users)))

	for _, userID := range users {
		token, _, err := s.issuer.Issue(ctx, pkgjwt.Identity{UserID: userID})
		if err != nil {
			return fmt.Errorf("issue token for user %s: %w", userID, err)
		}
		if err := s.smokeUser(ctx, userID, token); err != nil {
			return fmt.Errorf("smoke user %s: %w", userID, err)
		}
	}

	if err := s.writeAudit(ctx); err != nil {
		return fmt.Errorf("write smoke_ok audit: %w", err)
	}

	s.logger.Info("finance smoke complete", slog.Int("users_tested", len(users)))
	return nil
}

// smokeUser runs the RF-38e probe sequence for a single user.
func (s *financeSmoke) smokeUser(ctx context.Context, userID, token string) error {
	auth := "Bearer " + token

	// RF-38e – 1: paginated transaction listing
	if err := s.probeGET(ctx, "/api/v1/finance/transactions", auth); err != nil {
		return fmt.Errorf("list transactions: %w", err)
	}

	// RF-38e – 2: monthly summary
	if err := s.probeGET(ctx, "/api/v1/finance/summary", auth); err != nil {
		return fmt.Errorf("monthly summary: %w", err)
	}

	// RF-38e – 3: invoice listing + detail (decision E2: detail only when total > 0)
	body, err := s.probeGETBody(ctx, "/api/v1/finance/invoices", auth)
	if err != nil {
		return fmt.Errorf("list invoices: %w", err)
	}

	var invList smokeInvoiceList
	if jsonErr := json.Unmarshal(body, &invList); jsonErr == nil && invList.Total > 0 && len(invList.Items) > 0 {
		detailPath := "/api/v1/finance/invoices/" + invList.Items[0].ID
		if err := s.probeGET(ctx, detailPath, auth); err != nil {
			return fmt.Errorf("invoice detail: %w", err)
		}
	}

	s.logger.Info("smoke probe passed", slog.String("user_id", userID))
	return nil
}

// probeGET issues a GET request and requires a 2xx response.
func (s *financeSmoke) probeGET(ctx context.Context, path, authHeader string) error {
	_, err := s.probeGETBody(ctx, path, authHeader)
	return err
}

// probeGETBody issues a GET request, requires 2xx, and returns the response body bytes.
// ctx flows through http.NewRequestWithContext so future migrations away from
// fiber's app.Test (which currently ignores req context) inherit cancellation.
func (s *financeSmoke) probeGETBody(ctx context.Context, path, authHeader string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := s.http.Test(req)
	if err != nil {
		return nil, fmt.Errorf("test request %s: %w", path, err)
	}

	body, err := readBody(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body for %s: %w", path, err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("expected 2xx for %s, got %d: %s", path, resp.StatusCode, body)
	}
	return body, nil
}

func collectUserIDs(rows closableRows) (_ []string, err error) {
	defer closeInto(&err, rows)

	users := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		users = append(users, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func readBody(body io.ReadCloser) (_ []byte, err error) {
	defer closeInto(&err, body)

	content, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func closeInto(retErr *error, closer io.Closer) {
	if closeErr := closer.Close(); closeErr != nil && *retErr == nil {
		*retErr = closeErr
	}
}
