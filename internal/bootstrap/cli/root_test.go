package cli_test

import (
	"testing"

	bootstrapcli "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/cli"

	"github.com/stretchr/testify/require"
)

type runnerCall struct {
	name string
}

type fakeRunners struct {
	calls []runnerCall
}

func (f *fakeRunners) RunServer() error {
	f.calls = append(f.calls, runnerCall{name: "server"})
	return nil
}

func TestNewRootCommandRegistersNoLegacyCLISubcommands(t *testing.T) {
	cmd := bootstrapcli.NewRootCommand(&fakeRunners{})

	require.Equal(t, "financialcontrol-api", cmd.Use)
	// Planning/budget/sync CLI commands removed together with the planning module.
	require.Empty(t, cmd.Commands(), "no subcommands expected after legacy module removal")
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
