package database

import (
	"io/fs"
	"strings"
	"testing"
)

func TestMigrationsFS_ListsExactFiles(t *testing.T) {
	want := map[string]bool{
		"migrations/000000_baseline.up.sql":                 false,
		"migrations/000000_baseline.down.sql":               false,
		"migrations/000001_initial_schema.up.sql":           false,
		"migrations/000001_initial_schema.down.sql":         false,
		"migrations/000002_finance_module_ddl_dml.up.sql":   false,
		"migrations/000002_finance_module_ddl_dml.down.sql": false,
	}

	fsys := MigrationsFS()
	err := fs.WalkDir(fsys, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if _, ok := want[path]; ok {
			want[path] = true
		} else {
			t.Errorf("unexpected file in embed.FS: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	for path, found := range want {
		if !found {
			t.Errorf("expected file not found in embed.FS: %s", path)
		}
	}
}

func TestMigrationsFS_NoIfNotExists(t *testing.T) {
	fsys := MigrationsFS()
	data, err := fs.ReadFile(fsys, "migrations/000001_initial_schema.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(data), "IF NOT EXISTS") {
		t.Error("000001_initial_schema.up.sql must not contain 'IF NOT EXISTS'")
	}
	if strings.Contains(string(data), "DB_A453C8_FinancialControl.") {
		t.Error("000001_initial_schema.up.sql must not contain a physical database prefix")
	}
}
