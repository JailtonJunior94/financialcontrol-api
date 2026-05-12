package database

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

func TestMigrationsFS_ListsExactFiles(t *testing.T) {
	want := migrationFileSet(t, os.DirFS("."))
	got := migrationFileSet(t, MigrationsFS())

	if len(got) != len(want) {
		t.Fatalf("unexpected migration file count in embed.FS: got=%d want=%d", len(got), len(want))
	}
	for path := range want {
		if _, ok := got[path]; !ok {
			t.Errorf("expected file not found in embed.FS: %s", path)
		}
	}
	for path := range got {
		if _, ok := want[path]; !ok {
			t.Errorf("unexpected file in embed.FS: %s", path)
		}
	}
}

// BUG-007 regression: migration 000003 must use THROW (which aborts the batch
// unconditionally) instead of RAISERROR(...,16,1) (which does not by itself
// abort and would let the subsequent DROP TABLE statements run when the smoke
// gate has not been satisfied).
func TestMigrationsFS_000003UsesThrowForSmokeGate(t *testing.T) {
	fsys := MigrationsFS()
	data, err := fs.ReadFile(fsys, "migrations/000003_finance_module_drop_legacy.up.sql")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "THROW") {
		t.Error("000003_finance_module_drop_legacy.up.sql must use THROW for the smoke gate to abort the batch unconditionally")
	}
	if strings.Contains(content, "RAISERROR") {
		t.Error("000003_finance_module_drop_legacy.up.sql must not use RAISERROR for the smoke gate — severity 16 does not abort the batch")
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

func migrationFileSet(t *testing.T, fsys fs.FS) map[string]struct{} {
	t.Helper()

	files := make(map[string]struct{})
	err := fs.WalkDir(fsys, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sql") {
			files[path] = struct{}{}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	return files
}
