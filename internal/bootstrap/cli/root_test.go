package cli_test

import (
	"testing"
	"time"

	bootstrapcli "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/cli"

	"github.com/stretchr/testify/require"
)

type runnerCall struct {
	name     string
	date     time.Time
	category string
}

type fakeRunners struct {
	calls []runnerCall
}

func (f *fakeRunners) RunServer() error {
	f.calls = append(f.calls, runnerCall{name: "server"})
	return nil
}

func (f *fakeRunners) RunBudget(date time.Time) error {
	f.calls = append(f.calls, runnerCall{name: "budget", date: date})
	return nil
}

func (f *fakeRunners) RunBudgetCardsAndOthers(date time.Time) error {
	f.calls = append(f.calls, runnerCall{name: "budget-cards-and-others", date: date})
	return nil
}

func (f *fakeRunners) RunBudgetUnified(date time.Time) error {
	f.calls = append(f.calls, runnerCall{name: "budget-unified", date: date})
	return nil
}

func (f *fakeRunners) RunBudgetFull(date time.Time) error {
	f.calls = append(f.calls, runnerCall{name: "budget-full", date: date})
	return nil
}

func (f *fakeRunners) RunBalance(date time.Time) error {
	f.calls = append(f.calls, runnerCall{name: "balance", date: date})
	return nil
}

func (f *fakeRunners) RunBudgetByCategory(date time.Time, category string) error {
	f.calls = append(f.calls, runnerCall{name: "budget-category", date: date, category: category})
	return nil
}

func (f *fakeRunners) RunSync() error {
	f.calls = append(f.calls, runnerCall{name: "sync"})
	return nil
}

func TestNewRootCommandRegistersExpectedSubcommands(t *testing.T) {
	cmd := bootstrapcli.NewRootCommand(&fakeRunners{})

	require.Equal(t, "financialcontrol-api", cmd.Use)

	names := make([]string, 0, len(cmd.Commands()))
	for _, subcommand := range cmd.Commands() {
		names = append(names, subcommand.Name())
	}

	require.ElementsMatch(t, []string{
		"balance",
		"budget",
		"budget-cards-and-others",
		"budget-category",
		"budget-full",
		"budget-unified",
		"sync",
	}, names)
}

func TestNewRootCommandExecutesBudgetCategoryRunner(t *testing.T) {
	runners := &fakeRunners{}
	cmd := bootstrapcli.NewRootCommand(runners)
	cmd.SetArgs([]string{"budget-category", "--date", "17/04/2026", "--category", "Metas"})

	err := cmd.Execute()

	require.NoError(t, err)
	require.Len(t, runners.calls, 1)
	require.Equal(t, "budget-category", runners.calls[0].name)
	require.Equal(t, "Metas", runners.calls[0].category)
	require.Equal(t, time.Date(2026, time.April, 17, 0, 0, 0, 0, time.UTC), runners.calls[0].date)
}

func TestNewRootCommandExecutesServerByDefault(t *testing.T) {
	runners := &fakeRunners{}
	cmd := bootstrapcli.NewRootCommand(runners)
	cmd.SetArgs(nil)

	err := cmd.Execute()

	require.NoError(t, err)
	require.Len(t, runners.calls, 1)
	require.Equal(t, "server", runners.calls[0].name)
}
