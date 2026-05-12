//go:build integration

package mssql

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/microsoft/go-mssqldb"

	devkitmgr "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	devkitmssql "github.com/JailtonJunior94/devkit-go/pkg/database/mssql"
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
	sharedDB  *sql.DB
	sharedDSN string
	sharedErr error
	container testcontainers.Container
)

// GetSharedTestDatabase returns a shared *sql.DB backed by a SQL Server testcontainer.
// The container is started once per test binary via sync.Once.
// The returned cleanup func truncates all user tables in the test database.
func GetSharedTestDatabase() (*sql.DB, func(), error) {
	once.Do(func() {
		sharedDB, sharedDSN, sharedErr = startMSSQLContainer(context.Background())
	})
	if sharedErr != nil {
		return nil, nil, sharedErr
	}

	cleanup := func() {
		truncateAllTables(sharedDB)
	}

	return sharedDB, cleanup, nil
}

// GetSharedTestDSN returns the DSN of the shared SQL Server testcontainer.
// The container is started once per test binary via sync.Once.
func GetSharedTestDSN() (string, error) {
	once.Do(func() {
		sharedDB, sharedDSN, sharedErr = startMSSQLContainer(context.Background())
	})
	if sharedErr != nil {
		return "", sharedErr
	}
	return sharedDSN, nil
}

// GetSharedTestManager returns a devkit-go Manager connected to the shared SQL Server
// testcontainer. The container is started once per test binary via sync.Once.
// The returned cleanup func shuts down the manager and truncates all user tables.
func GetSharedTestManager() (devkitmgr.Manager, func(), error) {
	once.Do(func() {
		sharedDB, sharedDSN, sharedErr = startMSSQLContainer(context.Background())
	})
	if sharedErr != nil {
		return nil, nil, sharedErr
	}

	mgr, err := devkitmgr.New(devkitmssql.MSSQLConfig{DSN: sharedDSN})
	if err != nil {
		return nil, nil, fmt.Errorf("mssql testcontainer: manager: %w", err)
	}

	cleanup := func() {
		_ = mgr.Shutdown(context.Background())
		truncateAllTables(sharedDB)
	}

	return mgr, cleanup, nil
}

func startMSSQLContainer(ctx context.Context) (*sql.DB, string, error) {
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
		return nil, "", fmt.Errorf("mssql testcontainer: start: %w", err)
	}
	container = c

	host, err := c.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("mssql testcontainer: host: %w", err)
	}

	mappedPort, err := c.MappedPort(ctx, mssqlPort)
	if err != nil {
		return nil, "", fmt.Errorf("mssql testcontainer: port: %w", err)
	}

	dsn := fmt.Sprintf(
		"sqlserver://sa:%s@%s:%s?database=master&connection+timeout=30",
		mssqlSAPass, host, mappedPort.Port(),
	)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, "", fmt.Errorf("mssql testcontainer: open: %w", err)
	}

	for i := range connectMaxRetries {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = db.PingContext(pingCtx)
		cancel()
		if err == nil {
			break
		}
		if i < connectMaxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		_ = db.Close()
		return nil, "", fmt.Errorf("mssql testcontainer: connect: %w", err)
	}

	return db, dsn, nil
}

// truncateAllTables removes all rows from every user table in the test database.
// FK constraints are disabled during truncation to avoid ordering issues.
func truncateAllTables(db *sql.DB) {
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
