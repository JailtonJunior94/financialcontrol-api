package main

import (
	"log"

	"github.com/jailtonjunior94/financialcontrol-api/src/app"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "",
		Short: "financialcontrol-api",
		Run: func(cmd *cobra.Command, args []string) {
			app.RunServer()
		},
	}

	createTopics := &cobra.Command{
		Use:   "budget",
		Short: "Gera e atualiza orçamentos",
		Run: func(cmd *cobra.Command, args []string) {
			app.RunBudget()
		},
	}

	root.AddCommand(createTopics)
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
