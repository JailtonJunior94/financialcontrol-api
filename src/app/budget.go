package app

import (
	"os"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/olekukonko/tablewriter"
)

func RunBudget() {
	environments.SetupEnvironments()
	db := database.NewConnection()
	ioc.SetupDependencyInjection(db)

	data := [][]string{
		{"Janeiro", "R$ 14.530,00", "R$ 0,00", "Custos fixos", "R$ 200,00", "R$ 5.812,00", "10%", "15%"},
		{"Janeiro", "R$ 14.530,00", "R$ 0,00", "Conforto", "R$ 200,00", "R$ 5.812,00", "10%", "15%"},
		{"Janeiro", "R$ 14.530,00", "R$ 0,00", "Metas", "R$ 200,00", "R$ 5.812,00", "10%", "15%"},
		{"Janeiro", "R$ 14.530,00", "R$ 0,00", "Prazeres", "R$ 200,00", "R$ 5.812,00", "10%", "15%"},
		{"Janeiro", "R$ 14.530,00", "R$ 0,00", "Liberdade Financeira", "R$ 200,00", "R$ 5.812,00", "10%", "15%"},
		{"Janeiro", "R$ 14.530,00", "R$ 0,00", "Conhecimento", "R$ 200,00", "R$ 5.812,00", "10%", "15%"},
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Mês", "Total a Gastar", "Total Gastos", "Orçamento", "Valor Gasto [Fatura]", "Valor Gasto [Outros]", "Devo Gastar", "Utilizado", "Total"})

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}
