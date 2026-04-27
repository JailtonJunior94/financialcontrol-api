//go:build integration

package mssql

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const connectMaxRetries = 10

const (
	mssqlImage  = "mcr.microsoft.com/mssql/server:2022-latest"
	mssqlSAPass = "TestP@ss123!"
	mssqlPort   = "1433/tcp"
)

var (
	once      sync.Once
	sharedDB  *sqlx.DB
	sharedErr error
	container testcontainers.Container
)

// GetSharedTestDatabase returns a shared *sqlx.DB backed by a SQL Server testcontainer.
// The container is started once per test binary via sync.Once.
// The returned cleanup func truncates all user tables in the test database.
func GetSharedTestDatabase() (*sqlx.DB, func(), error) {
	once.Do(func() {
		sharedDB, sharedErr = startMSSQLContainer(context.Background())
	})
	if sharedErr != nil {
		return nil, nil, sharedErr
	}

	cleanup := func() {
		truncateAllTables(sharedDB)
	}

	return sharedDB, cleanup, nil
}

func startMSSQLContainer(ctx context.Context) (*sqlx.DB, error) {
	req := testcontainers.ContainerRequest{
		Image:        mssqlImage,
		ExposedPorts: []string{mssqlPort},
		Env: map[string]string{
			"ACCEPT_EULA": "Y",
			"SA_PASSWORD": mssqlSAPass,
			"MSSQL_PID":   "Developer",
		},
		WaitingFor: wait.ForLog("SQL Server is now ready for client connections").
			WithStartupTimeout(120 * time.Second),
	}

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("mssql testcontainer: start: %w", err)
	}
	container = c

	host, err := c.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("mssql testcontainer: host: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, mssqlPort)
	if err != nil {
		return nil, fmt.Errorf("mssql testcontainer: port: %w", err)
	}

	dsn := fmt.Sprintf(
		"sqlserver://sa:%s@%s:%s?database=master&connection+timeout=30",
		mssqlSAPass, host, mappedPort.Port(),
	)

	var db *sqlx.DB
	for i := range connectMaxRetries {
		db, err = sqlx.Connect("sqlserver", dsn)
		if err == nil {
			break
		}
		if i < connectMaxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("mssql testcontainer: connect: %w", err)
	}

	return db, nil
}

// truncateAllTables removes all rows from every user table in the test database.
// FK constraints are disabled during truncation to avoid ordering issues.
func truncateAllTables(db *sqlx.DB) {
	if db == nil {
		return
	}
	ctx := context.Background()

	var tables []string
	rows, err := db.QueryContext(ctx, `
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_TYPE = 'BASE TABLE'
		  AND TABLE_CATALOG = DB_NAME()
	`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tables = append(tables, t)
		}
	}

	for _, t := range tables {
		_, _ = db.ExecContext(ctx, fmt.Sprintf("DELETE FROM [%s]", t))
	}
}
