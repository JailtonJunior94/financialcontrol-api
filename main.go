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

	budgetUnified := &cobra.Command{
		Use:   "budget-unified",
		Short: "Visão unificada: orçamento, cartão e outros em uma tabela",
		Run: func(cmd *cobra.Command, args []string) {
			var date time.Time

			if dateParam == "" {
				date = time.Now()
			}

			if dateParam != "" {
				date, _ = time.Parse("02/01/2006", dateParam)
			}

			app.RunBudgetUnified(date)
		},
	}
	budgetUnified.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")

	budgetFull := &cobra.Command{
		Use:   "budget-full",
		Short: "Visão completa do orçamento com detalhamento por cartão e outros",
		Run: func(cmd *cobra.Command, args []string) {
			var date time.Time

			if dateParam == "" {
				date = time.Now()
			}

			if dateParam != "" {
				date, _ = time.Parse("02/01/2006", dateParam)
			}

			app.RunBudgetFullView(date)
		},
	}
	budgetFull.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")

	balance := &cobra.Command{
		Use:   "balance",
		Short: "Saldo disponível real por categoria e detalhamento do cartão por subcategoria",
		Run: func(cmd *cobra.Command, args []string) {
			var date time.Time

			if dateParam == "" {
				date = time.Now()
			}

			if dateParam != "" {
				date, _ = time.Parse("02/01/2006", dateParam)
			}

			app.RunBalance(date)
		},
	}
	balance.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")

	var categoryParam string

	budgetByCategory := &cobra.Command{
		Use:   "budget-category",
		Short: "Lista todos os itens de uma categoria com valores e data da compra",
		Run: func(cmd *cobra.Command, args []string) {
			var date time.Time

			if dateParam == "" {
				date = time.Now()
			}

			if dateParam != "" {
				date, _ = time.Parse("02/01/2006", dateParam)
			}

			if categoryParam == "" {
				log.Fatal("a flag --category é obrigatória")
			}

			app.RunBudgetByCategory(date, categoryParam)
		},
	}
	budgetByCategory.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	budgetByCategory.Flags().StringVarP(&categoryParam, "category", "c", "", "Categoria para filtrar (obrigatório)")

	root.AddCommand(budget, budgetCardsAndOthers, budgetUnified, budgetFull, balance, budgetByCategory, sync)
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
