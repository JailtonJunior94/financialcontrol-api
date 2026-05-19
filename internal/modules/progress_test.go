package modules_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RefactorProgressSnapshot holds measurable indicators for the modularization
// migration as defined in techspec.md (Monitoramento e Observabilidade).
type RefactorProgressSnapshot struct {
	ModuleName              string
	LegacyArtifactsCount    int
	ModularArtifactsCount   int
	CrossModuleDepsDetected int
}

// legacyCapabilityDirs are the layer-oriented directories that should shrink
// as capabilities migrate to internal/modules/<name>.
var legacyCapabilityDirs = []string{
	"internal/bootstrap/cli",
	"internal/application/services",
	"internal/application/handlers",
	"internal/application/mappings",
	"internal/application/usecase",
	"internal/domain/usecases",
	"internal/domain/interfaces",
	"internal/http/controllers",
	"internal/http/routes",
	"internal/infrastructure/repositories",
	"internal/infrastructure/queries",
}

// moduleDirs are the module roots that should grow as migration progresses.
var moduleDirs = []string{
	"internal/modules/identity",
	"internal/modules/cards",
	"internal/modules/categories",
	"internal/modules/finance",
}

// crossModuleRestricted are import paths that must not appear in modular
// packages. All active modules are covered following the completed migration.
var crossModuleRestricted = map[string][]string{
	"identity":   {"internal/infrastructure/repositories", "internal/application/handlers"},
	"cards":      {"internal/infrastructure/repositories", "internal/application/handlers"},
	"categories": {"internal/infrastructure/repositories", "internal/application/handlers"},
	"finance":    {"internal/infrastructure/repositories", "internal/application/handlers"},
}

func TestRefactorProgress_Snapshot(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	legacy := countGoFiles(t, repoRoot, legacyCapabilityDirs)
	modular := countGoFiles(t, repoRoot, moduleDirs)

	snapshot := RefactorProgressSnapshot{
		ModuleName:            "all",
		LegacyArtifactsCount:  legacy,
		ModularArtifactsCount: modular,
	}

	t.Logf("refactor progress snapshot: legacy_artifacts=%d modular_artifacts=%d",
		snapshot.LegacyArtifactsCount, snapshot.ModularArtifactsCount)

	// Modular artifacts must already exceed legacy capability artifacts —
	// this gate confirms the migration is past its midpoint.
	assert.Greater(t, snapshot.ModularArtifactsCount, snapshot.LegacyArtifactsCount,
		"modular artifacts must outnumber legacy capability artifacts")
}

func TestRefactorProgress_CrossModuleDependencies(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	violations := 0

	for module, restricted := range crossModuleRestricted {
		modulePath := filepath.Join(repoRoot, "internal", "modules", module)
		if err := filepath.WalkDir(modulePath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			imports := parseFileImports(t, path)
			for _, imp := range imports {
				for _, blocked := range restricted {
					if strings.Contains(imp, blocked) {
						t.Logf("cross-module dependency detected: %s imports %s (blocked: %s)", path, imp, blocked)
						violations++
					}
				}
			}
			return nil
		}); err != nil {
			t.Logf("walk error for module %s: %v", module, err)
		}
	}

	t.Logf("cross-module dependency violations detected: %d", violations)
	assert.Equal(t, 0, violations, "modular packages must not import legacy infrastructure repositories or handlers")
}

func TestRefactorProgress_ModuleRoutesRegisteredFromModules(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	// Verify that the pkg/http adapter delegates route registration to
	// modules rather than importing route files from internal/http/routes.
	// The bootstrap/http layer delegates to pkg/http, which is the actual
	// composition point for module-level route registration.
	platformHTTP := filepath.Join(repoRoot, "pkg", "http", "router.go")
	imports := parseFileImports(t, platformHTTP)

	usesModuleRegistry := false
	for _, imp := range imports {
		if strings.Contains(imp, "internal/modules") {
			usesModuleRegistry = true
			break
		}
	}

	t.Logf("platform/http imports: %v", imports)
	assert.True(t, usesModuleRegistry, "platform/http must register routes through internal/modules registry")
}

func TestRefactorProgress_LegacyEquivalentArtifactsRemoved(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	testCases := []struct {
		name        string
		relative    string
		shouldExist bool
		goFilesOnly bool
	}{
		{
			name:        "legacy HTTP registrars removed",
			relative:    "internal/http/routes/register.go",
			shouldExist: false,
		},
		{
			name:        "legacy route package removed",
			relative:    "internal/http/routes",
			shouldExist: false,
			goFilesOnly: true,
		},
		{
			name:        "legacy controller package removed",
			relative:    "internal/http/controllers",
			shouldExist: false,
			goFilesOnly: true,
		},
		{
			name:        "legacy budget CLI entrypoints removed",
			relative:    "internal/bootstrap/cli/budget.go",
			shouldExist: false,
		},
		{
			name:        "legacy sync CLI entrypoint removed",
			relative:    "internal/bootstrap/cli/sync.go",
			shouldExist: false,
		},
		{
			name:        "modular platform router preserved",
			relative:    "pkg/http/router.go",
			shouldExist: true,
		},
		{
			name:        "planning module removed",
			relative:    "internal/modules/planning",
			shouldExist: false,
			goFilesOnly: true,
		},
		{
			name:        "billing module removed",
			relative:    "internal/modules/billing",
			shouldExist: false,
			goFilesOnly: true,
		},
		{
			name:        "transactions module removed",
			relative:    "internal/modules/transactions",
			shouldExist: false,
			goFilesOnly: true,
		},
		{
			name:        "invoicing module removed",
			relative:    "internal/modules/invoicing",
			shouldExist: false,
			goFilesOnly: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			target := filepath.Join(repoRoot, tc.relative)
			_, err := os.Stat(target)
			if tc.shouldExist {
				assert.NoError(t, err, "expected %s to exist", tc.relative)
				return
			}

			if tc.goFilesOnly {
				assert.Zero(t, countGoFiles(t, repoRoot, []string{tc.relative}), "expected %s to have no legacy Go files", tc.relative)
				return
			}

			assert.ErrorIs(t, err, os.ErrNotExist, "expected %s to be removed", tc.relative)
		})
	}
}

// countGoFiles counts non-test .go files under the given relative directories.
func countGoFiles(t *testing.T, root string, dirs []string) int {
	t.Helper()
	count := 0
	for _, dir := range dirs {
		_ = filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				count++
			}
			return nil
		})
	}
	return count
}

// parseFileImports returns the import paths declared in a Go source file.
func parseFileImports(t *testing.T, absPath string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, absPath, nil, parser.ImportsOnly)
	if err != nil {
		t.Logf("warning: could not parse %s: %v", absPath, err)
		return nil
	}

	imports := make([]string, 0, len(file.Imports))
	for _, imp := range file.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, `"`))
	}
	return imports
}
