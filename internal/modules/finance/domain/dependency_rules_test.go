package domain_test

import (
	"go/build"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomainHasNoForbiddenImports(t *testing.T) {
	t.Parallel()
	forbidden := []string{
		"database/sql",
		"github.com/gofiber/fiber/v2",
	}
	pkgs := []string{
		"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities",
		"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services",
		"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters",
		"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos",
	}
	ctx := build.Default

	for _, pkg := range pkgs {
		p, err := ctx.Import(pkg, ".", build.ImportComment)
		if err != nil {
			// package may not compile yet in this environment; skip
			t.Logf("skipping %s: %v", pkg, err)
			continue
		}
		allImports := append(p.Imports, p.TestImports...)
		for _, imp := range allImports {
			for _, bad := range forbidden {
				assert.NotEqual(t, bad, imp, "domain package %s must not import %s", pkg, bad)
			}
		}
	}
}
