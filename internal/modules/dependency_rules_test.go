package modules_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allModules = []string{
	"identity", "categories", "cards", "billing",
	"transactions", "invoicing", "planning",
}

func TestContainerWiresInvoiceEventThroughModuleContracts(t *testing.T) {
	imports := fileImports(t, "internal/bootstrap/container/container.go")

	assert.NotContains(t, imports, "github.com/jailtonjunior94/financialcontrol-api/internal/application/handlers")
	assert.Contains(t, imports, "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application")
	assert.Contains(t, imports, "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application")
}

func TestModularContractsAvoidCrossModuleInfrastructureImports(t *testing.T) {
	restrictedImports := []string{
		"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/repositories",
		"github.com/jailtonjunior94/financialcontrol-api/internal/application/handlers",
	}

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	for _, mod := range allModules {
		modPath := filepath.Join(repoRoot, "internal", "modules", mod)
		if err := filepath.WalkDir(modPath, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			imports := parseFileImports(t, path)
			for _, imp := range imports {
				for _, restricted := range restrictedImports {
					assert.NotContains(t, imp, restricted,
						"module %s must not import legacy infrastructure (%s): file %s", mod, restricted, path)
				}
			}
			return nil
		}); err != nil {
			t.Logf("walk error for module %s: %v", mod, err)
		}
	}
}

// TestNoCrossModuleDomainImports verifies that no module's .go files import the
// domain/ package of another module (RF-12).
func TestNoCrossModuleDomainImports(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	for _, mod := range allModules {
		t.Run(mod, func(t *testing.T) {
			modPath := filepath.Join(repoRoot, "internal", "modules", mod)
			if err := filepath.WalkDir(modPath, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				imports := parseFileImports(t, path)
				for _, imp := range imports {
					for _, other := range allModules {
						if other == mod {
							continue
						}
						crossDomainPkg := fmt.Sprintf("internal/modules/%s/domain", other)
						assert.NotContains(t, imp, crossDomainPkg,
							"module %s must not import domain/ of module %s: file %s", mod, other, path)
					}
				}
				return nil
			}); err != nil {
				t.Logf("walk error for module %s: %v", mod, err)
			}
		})
	}
}

// TestNoCrossModuleInfrastructureImports verifies that no module's .go files
// import the infrastructure/ package of another module (RF-13).
func TestNoCrossModuleInfrastructureImports(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	for _, mod := range allModules {
		t.Run(mod, func(t *testing.T) {
			modPath := filepath.Join(repoRoot, "internal", "modules", mod)
			if err := filepath.WalkDir(modPath, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				imports := parseFileImports(t, path)
				for _, imp := range imports {
					for _, other := range allModules {
						if other == mod {
							continue
						}
						crossInfraPkg := fmt.Sprintf("internal/modules/%s/infrastructure", other)
						assert.NotContains(t, imp, crossInfraPkg,
							"module %s must not import infrastructure/ of module %s: file %s", mod, other, path)
					}
				}
				return nil
			}); err != nil {
				t.Logf("walk error for module %s: %v", mod, err)
			}
		})
	}
}

// TestNoLegacyDomainImports verifies that no module imports from the legacy
// internal/domain/ packages (RF-12, RF-13).
func TestNoLegacyDomainImports(t *testing.T) {
	legacyPkgs := []string{
		"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities",
		"github.com/jailtonjunior94/financialcontrol-api/internal/domain/customerrors",
		"github.com/jailtonjunior94/financialcontrol-api/internal/domain/constants",
	}

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	for _, mod := range allModules {
		t.Run(mod, func(t *testing.T) {
			modPath := filepath.Join(repoRoot, "internal", "modules", mod)
			if err := filepath.WalkDir(modPath, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				imports := parseFileImports(t, path)
				for _, imp := range imports {
					for _, legacy := range legacyPkgs {
						assert.NotEqual(t, imp, legacy,
							"module %s must not import legacy package %s: file %s", mod, legacy, path)
					}
				}
				return nil
			}); err != nil {
				t.Logf("walk error for module %s: %v", mod, err)
			}
		})
	}
}

// TestNoLegacyPlatformImports verifies that no module imports from the legacy
// internal/platform/ or internal/shared/ packages (RF-15, RF-16).
func TestNoLegacyPlatformImports(t *testing.T) {
	legacyPrefixes := []string{
		"github.com/jailtonjunior94/financialcontrol-api/internal/platform/",
		"github.com/jailtonjunior94/financialcontrol-api/internal/shared/",
	}

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	repoRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")

	for _, mod := range allModules {
		t.Run(mod, func(t *testing.T) {
			modPath := filepath.Join(repoRoot, "internal", "modules", mod)
			if err := filepath.WalkDir(modPath, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				imports := parseFileImports(t, path)
				for _, imp := range imports {
					for _, prefix := range legacyPrefixes {
						assert.False(t, strings.HasPrefix(imp, prefix),
							"module %s must not import legacy package %s: file %s", mod, imp, path)
					}
				}
				return nil
			}); err != nil {
				t.Logf("walk error for module %s: %v", mod, err)
			}
		})
	}
}

func fileImports(t *testing.T, relativePath string) []string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	fset := token.NewFileSet()
	absPath := filepath.Join(filepath.Dir(currentFile), "..", "..", relativePath)
	file, err := parser.ParseFile(fset, absPath, nil, parser.ImportsOnly)
	require.NoError(t, err)

	imports := make([]string, 0, len(file.Imports))
	for _, imported := range file.Imports {
		imports = append(imports, strings.Trim(imported.Path.Value, `"`))
	}

	return imports
}
