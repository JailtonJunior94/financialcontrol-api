package modules_test

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const moduleImportPrefix = "github.com/jailtonjunior94/financialcontrol-api/internal/modules/"

// moduleNameFromImport returns the bounded-context module name for an import
// path under internal/modules/<name>/..., or "" when the path is not a module
// subpackage (e.g. the internal/modules registry package itself).
func moduleNameFromImport(imp string) string {
	if !strings.HasPrefix(imp, moduleImportPrefix) {
		return ""
	}
	rest := strings.TrimPrefix(imp, moduleImportPrefix)
	if rest == "" {
		return ""
	}
	if name, _, found := strings.Cut(rest, "/"); found {
		return name
	}
	return rest
}

func repoRootFromCaller(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(currentFile), "..", "..")
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// discoverModules returns the immediate subdirectories of internal/modules that
// contain at least one non-test .go file — i.e. real bounded-context modules,
// not the registry package or its test files.
func discoverModules(t *testing.T, repoRoot string) []string {
	t.Helper()
	modulesDir := filepath.Join(repoRoot, "internal", "modules")
	entries, err := os.ReadDir(modulesDir)
	require.NoError(t, err)

	var mods []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		hasProductionGo := false
		_ = filepath.WalkDir(filepath.Join(modulesDir, e.Name()), func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil || d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				hasProductionGo = true
			}
			return nil
		})
		if hasProductionGo {
			mods = append(mods, e.Name())
		}
	}
	require.NotEmpty(t, mods, "expected at least one module under internal/modules")
	return mods
}

// TestOnlyContainerImportsMultipleModules enforces that the composition root
// (internal/bootstrap/container) is the sole place allowed to import more than
// one bounded-context module. This guards the modular-monolith boundary against
// accidental cross-module coupling outside the wiring layer.
//
// Documented exception: internal/modules/<m>/infrastructure/providers/ holds
// intentional Ports & Adapters cross-module adapters — the same exception
// already carved out by TestNoCrossModuleDomainImports.
func TestOnlyContainerImportsMultipleModules(t *testing.T) {
	repoRoot := repoRootFromCaller(t)

	containerPrefix := filepath.Join("internal", "bootstrap", "container") + string(filepath.Separator)
	providersFragment := string(filepath.Separator) + filepath.Join("infrastructure", "providers") + string(filepath.Separator)

	// Only Go source roots can hold multi-module imports; restricting the walk
	// keeps the check deterministic and avoids scanning skill/doc scaffolding.
	for _, root := range []string{"internal", "pkg", "cmd"} {
		walkRoot := filepath.Join(repoRoot, root)
		err := filepath.WalkDir(walkRoot, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if d.IsDir() {
				if d.Name() == "vendor" || d.Name() == "testdata" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, relErr := filepath.Rel(repoRoot, path)
			if relErr != nil {
				return nil
			}
			if strings.HasPrefix(rel, containerPrefix) {
				return nil
			}
			if strings.Contains(string(filepath.Separator)+rel, providersFragment) {
				return nil
			}

			seen := map[string]struct{}{}
			for _, imp := range parseFileImports(t, path) {
				if m := moduleNameFromImport(imp); m != "" {
					seen[m] = struct{}{}
				}
			}
			assert.LessOrEqualf(t, len(seen), 1,
				"only internal/bootstrap/container may import multiple modules; %s imports %d modules: %v",
				rel, len(seen), sortedKeys(seen))
			return nil
		})
		require.NoError(t, err)
	}
}

// TestEveryModuleHasEntrypointAndIsWired guards against orphaned, unreachable
// modules (the planning-module regression: it lived for months with logic but
// no module.go and no container wiring, undetected). Every bounded-context
// module must expose a module.go composition entrypoint AND be referenced by
// the container.
func TestEveryModuleHasEntrypointAndIsWired(t *testing.T) {
	repoRoot := repoRootFromCaller(t)
	mods := discoverModules(t, repoRoot)

	wired := map[string]struct{}{}
	containerFile := filepath.Join(repoRoot, "internal", "bootstrap", "container", "container.go")
	for _, imp := range parseFileImports(t, containerFile) {
		if m := moduleNameFromImport(imp); m != "" {
			wired[m] = struct{}{}
		}
	}

	for _, mod := range mods {
		t.Run(mod, func(t *testing.T) {
			moduleGo := filepath.Join(repoRoot, "internal", "modules", mod, "module.go")
			_, statErr := os.Stat(moduleGo)
			assert.NoErrorf(t, statErr,
				"module %q must expose internal/modules/%s/module.go as its composition entrypoint", mod, mod)

			_, isWired := wired[mod]
			assert.Truef(t, isWired,
				"module %q is not referenced by internal/bootstrap/container/container.go — orphaned/unreachable module", mod)
		})
	}
}
