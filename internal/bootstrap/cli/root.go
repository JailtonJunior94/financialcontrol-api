package cli

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules"
	platformmodules "github.com/jailtonjunior94/financialcontrol-api/pkg/modules"
	"github.com/spf13/cobra"
)

type Runners = platformmodules.CLIRunner

func Execute() error {
	r, err := NewAppRunners()
	if err != nil {
		return err
	}
	return NewRootCommand(r).Execute()
}

func NewRootCommand(r Runners) *cobra.Command {
	root := &cobra.Command{
		Use:   "financialcontrol-api",
		Short: "financialcontrol-api",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunServer()
		},
	}

	modules.RegisterCLI(root, r)

	return root
}
