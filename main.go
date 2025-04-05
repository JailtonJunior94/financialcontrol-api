package main

import (
	"log"
	"time"

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

	var dateParam string

	budget := &cobra.Command{
		Use:   "budget",
		Short: "Gera e atualiza orçamentos",
		Run: func(cmd *cobra.Command, args []string) {
			var date time.Time

			if dateParam == "" {
				date = time.Now()
			}

			if dateParam != "" {
				date, _ = time.Parse("02/01/2006", dateParam)
			}

			app.RunBudget(date)
		},
	}
	budget.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")

	budgetCardsAndOthers := &cobra.Command{
		Use:   "budget-cards-and-others",
		Short: "Gera e atualiza orçamentos",
		Run: func(cmd *cobra.Command, args []string) {
			var date time.Time

			if dateParam == "" {
				date = time.Now()
			}

			if dateParam != "" {
				date, _ = time.Parse("02/01/2006", dateParam)
			}

			app.RunBudgetCardAndOthers(date)
		},
	}
	budgetCardsAndOthers.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")

	sync := &cobra.Command{
		Use:   "sync",
		Short: "Sincroniza os dados",
		Run: func(cmd *cobra.Command, args []string) {
			app.RunSync()
		},
	}

	root.AddCommand(budget, budgetCardsAndOthers, sync)
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
