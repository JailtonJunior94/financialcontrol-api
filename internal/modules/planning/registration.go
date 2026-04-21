package planning

import (
	"errors"
	"time"

	platformmodules "github.com/jailtonjunior94/financialcontrol-api/internal/platform/modules"

	"github.com/spf13/cobra"
)

func Registration() platformmodules.ModuleRegistration {
	return platformmodules.ModuleRegistration{
		Name: "planning",
		RegisterCLI: func(root *cobra.Command, runners platformmodules.CLIRunner) {
			root.AddCommand(
				newBudgetCommand(runners),
				newBudgetCardsAndOthersCommand(runners),
				newBudgetUnifiedCommand(runners),
				newBudgetFullCommand(runners),
				newBalanceCommand(runners),
				newBudgetCategoryCommand(runners),
				newSyncCommand(runners),
			)
		},
	}
}

func newBudgetCommand(r platformmodules.CLIRunner) *cobra.Command {
	var dateParam string

	cmd := &cobra.Command{
		Use:   "budget",
		Short: "Gera e atualiza orçamentos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunBudget(parseDate(dateParam))
		},
	}
	cmd.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	return cmd
}

func newBudgetCardsAndOthersCommand(r platformmodules.CLIRunner) *cobra.Command {
	var dateParam string

	cmd := &cobra.Command{
		Use:   "budget-cards-and-others",
		Short: "Gera e atualiza orçamentos",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunBudgetCardsAndOthers(parseDate(dateParam))
		},
	}
	cmd.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	return cmd
}

func newBudgetUnifiedCommand(r platformmodules.CLIRunner) *cobra.Command {
	var dateParam string

	cmd := &cobra.Command{
		Use:   "budget-unified",
		Short: "Visão unificada: orçamento, cartão e outros em uma tabela",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunBudgetUnified(parseDate(dateParam))
		},
	}
	cmd.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	return cmd
}

func newBudgetFullCommand(r platformmodules.CLIRunner) *cobra.Command {
	var dateParam string

	cmd := &cobra.Command{
		Use:   "budget-full",
		Short: "Visão completa do orçamento com detalhamento por cartão e outros",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunBudgetFull(parseDate(dateParam))
		},
	}
	cmd.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	return cmd
}

func newBalanceCommand(r platformmodules.CLIRunner) *cobra.Command {
	var dateParam string

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Saldo disponível real por categoria e detalhamento do cartão por subcategoria",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunBalance(parseDate(dateParam))
		},
	}
	cmd.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	return cmd
}

func newBudgetCategoryCommand(r platformmodules.CLIRunner) *cobra.Command {
	var dateParam string
	var categoryParam string

	cmd := &cobra.Command{
		Use:   "budget-category",
		Short: "Lista todos os itens de uma categoria com valores e data da compra",
		RunE: func(cmd *cobra.Command, args []string) error {
			if categoryParam == "" {
				return errors.New("a flag --category é obrigatória")
			}

			return r.RunBudgetByCategory(parseDate(dateParam), categoryParam)
		},
	}
	cmd.Flags().StringVarP(&dateParam, "date", "d", "", "Data no formato DD/MM/YYYY (opcional)")
	cmd.Flags().StringVarP(&categoryParam, "category", "c", "", "Categoria para filtrar (obrigatório)")
	return cmd
}

func newSyncCommand(r platformmodules.CLIRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sincroniza os dados",
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunSync()
		},
	}
}

func parseDate(value string) time.Time {
	if value == "" {
		return time.Now()
	}

	date, err := time.Parse("02/01/2006", value)
	if err != nil {
		return time.Time{}
	}

	return date
}
