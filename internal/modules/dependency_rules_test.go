package modules_test

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	for _, relativePath := range []string{
		"internal/modules/invoicing/application/invoice_changed_handler.go",
		"internal/modules/planning/ports.go",
	} {
		imports := fileImports(t, relativePath)
		for _, restrictedImport := range restrictedImports {
			assert.NotContains(t, imports, restrictedImport, relativePath)
		}
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
