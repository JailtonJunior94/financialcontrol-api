package app

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/ioc"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/leekchan/accounting"
	"github.com/olekukonko/tablewriter"
)

func RunBudget(dateParam time.Time) {
	environments.SetupEnvironments()
	db := database.NewConnection()
	ioc.SetupDependencyInjection(db)

	date := time.Date(dateParam.Year(), dateParam.Month(), 1, 0, 0, 0, 0, time.UTC)

	budget := entities.NewBudget(date, vos.NewMoney(14_305.05))
	budgetMetas := entities.NewBudgetItem(budget, "Metas", vos.NewPercentage(0.10))
	bugetConforto := entities.NewBudgetItem(budget, "Conforto", vos.NewPercentage(0.10))
	budgetPrazeres := entities.NewBudgetItem(budget, "Prazeres", vos.NewPercentage(0.15))
	budgetCustoFixos := entities.NewBudgetItem(budget, "Custos fixos", vos.NewPercentage(0.40))
	budgetConhecimento := entities.NewBudgetItem(budget, "Conhecimento", vos.NewPercentage(0.05))
	budgetLiberdadeFinanceira := entities.NewBudgetItem(budget, "Liberdade Financeira", vos.NewPercentage(0.20))

	budget.AddItems([]*entities.BudgetItem{
		budgetCustoFixos,
		bugetConforto,
		budgetMetas,
		budgetPrazeres,
		budgetConhecimento,
		budgetLiberdadeFinanceira,
	})

	budgetLiberdadeFinanceira.AddAmountUsed(vos.NewMoney(2_861.01))

	invoices, err := ioc.InvoiceRepository.GetInvoices(date.AddDate(0, 1, 0))
	if err != nil {
		log.Fatalf("error fetching invoices: %v", err)
	}

	bills, err := ioc.BillRepository.Get(date)
	if err != nil {
		log.Fatalf("error fetching bills: %v", err)
	}

	conforto := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return strings.Contains(b.Description, "Faxina")
	})

	confortoSum := linq.Sum(conforto, func(i dtos.BillItemQuery) float64 {
		return i.Total
	})
	bugetConforto.AddAmountUsed(vos.NewMoney(confortoSum))

	custoFixo := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return !strings.Contains(b.Description, "Faxina")
	})

	custoFixoSum := linq.Sum(custoFixo, func(i dtos.BillItemQuery) float64 {
		return i.Total
	})
	budgetCustoFixos.AddAmountUsed(vos.NewMoney(custoFixoSum))

	groupedByTag := linq.GroupBy(invoices.Items, func(i dtos.InvoiceItemRead) string {
		return i.Tags
	})

	for tag, group := range groupedByTag {
		sum := linq.Sum(group, func(i dtos.InvoiceItemRead) float64 {
			return i.InstallmentValue
		})

		for _, budget := range budget.Items {
			if budget.Category == tag {
				budget.AddAmountUsed(vos.NewMoney(sum))
			}
		}
	}

	data := [][]string{}
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Mês", "Total a Gastar (Mensal)", "Orçamento", "Orçamento %", "Devo Gastar", "Valor Gasto", "Ainda Posso Gastar"})

	for _, item := range budget.Items {
		data = append(data, []string{
			budget.Date.Format("January 2006"),
			formatMoney(budget.AmountGoal),
			item.Category,
			formatPercentage(item.PercentageGoal),
			formatMoney(item.AmountGoal),
			formatMoney(item.AmountUsed),
			formatMoney(item.AmountGoal.Sub(item.AmountUsed)),
		})
	}

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}

func RunBudgetCardAndOthers(dateParam time.Time) {
	environments.SetupEnvironments()
	db := database.NewConnection()
	ioc.SetupDependencyInjection(db)

	date := time.Date(dateParam.Year(), dateParam.Month(), 1, 0, 0, 0, 0, time.UTC)

	invoices, err := ioc.InvoiceRepository.GetInvoices(date.AddDate(0, 1, 0))
	if err != nil {
		log.Fatalf("error fetching invoices: %v", err)
	}

	bills, err := ioc.BillRepository.Get(date)
	if err != nil {
		log.Fatalf("error fetching bills: %v", err)
	}

	conforto := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return strings.Contains(b.Description, "Faxina")
	})

	confortoSoma := linq.Sum(conforto, func(i dtos.BillItemQuery) float64 {
		return i.Total
	})

	custoFixo := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return !strings.Contains(b.Description, "Faxina")
	})

	custoFixoSoma := linq.Sum(custoFixo, func(i dtos.BillItemQuery) float64 {
		return i.Total
	})

	groupedByTag := linq.GroupBy(invoices.Items, func(i dtos.InvoiceItemRead) string {
		return i.Tags
	})

	gastoPrazeres := &BudgetCardAndOthers{Date: date, Category: "Prazeres", TotalSpentOthers: vos.NewMoney(0)}
	gastoConhecimento := &BudgetCardAndOthers{Date: date, Category: "Conhecimento", TotalSpentOthers: vos.NewMoney(0)}
	gastoConforto := &BudgetCardAndOthers{Date: date, Category: "Conforto", TotalSpentOthers: vos.NewMoney(confortoSoma)}
	gastoCustoFixos := &BudgetCardAndOthers{Date: date, Category: "Custos fixos", TotalSpentOthers: vos.NewMoney(custoFixoSoma)}

	BudgetCard := &BudgetCard{
		Items: []*BudgetCardAndOthers{
			gastoConforto,
			gastoPrazeres,
			gastoCustoFixos,
			gastoConhecimento,
		},
	}

	for tag, group := range groupedByTag {
		sum := linq.Sum(group, func(i dtos.InvoiceItemRead) float64 {
			return i.InstallmentValue
		})

		for _, budget := range BudgetCard.Items {
			if budget.Category == tag {
				budget.AddAmountUsed(vos.NewMoney(sum))
			}
		}
	}

	data := [][]string{}
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Mês", "Categoria", "Total Gasto (Cartão)", "Total Gasto (Outros)", "Total"})

	for _, item := range BudgetCard.Items {
		data = append(data, []string{
			item.Date.Format("January 2006"),
			item.Category,
			formatMoney(item.TotalSpentCard),
			formatMoney(item.TotalSpentOthers),
			formatMoney(item.Total),
		})
	}

	for _, v := range data {
		table.Append(v)
	}

	table.Render()
}

type BudgetCard struct {
	Items []*BudgetCardAndOthers
}

type BudgetCardAndOthers struct {
	Date             time.Time
	Category         string
	TotalSpentCard   vos.Money
	TotalSpentOthers vos.Money
	Total            vos.Money
}

func (b *BudgetCardAndOthers) AddAmountUsed(amount vos.Money) {
	b.TotalSpentCard = b.TotalSpentCard.Add(amount)
	b.Total = b.TotalSpentCard.Add(b.TotalSpentOthers)
}

func formatMoney(value vos.Money) string {
	format := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ".", Decimal: ","}
	return format.FormatMoney(value.Money())
}

func formatPercentage(value vos.Percentage) string {
	return fmt.Sprintf("%.2f%%", value.Percentage()*100)
}
