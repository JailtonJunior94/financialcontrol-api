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

	budget := entities.NewBudget(date, vos.NewMoney(13_874.40))
	budgetMetas := entities.NewBudgetItem(budget, "Metas", vos.NewPercentage(0.15))
	bugetConforto := entities.NewBudgetItem(budget, "Conforto", vos.NewPercentage(0.10))
	budgetPrazeres := entities.NewBudgetItem(budget, "Prazeres", vos.NewPercentage(0.10))
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

	budgetLiberdadeFinanceira.AddAmountUsed(vos.NewMoney(2_774.88))

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

func RunBudgetUnified(dateParam time.Time) {
	environments.SetupEnvironments()
	db := database.NewConnection()
	ioc.SetupDependencyInjection(db)

	date := time.Date(dateParam.Year(), dateParam.Month(), 1, 0, 0, 0, 0, time.UTC)

	budget := entities.NewBudget(date, vos.NewMoney(13_874.40))
	budgetMetas := entities.NewBudgetItem(budget, "Metas", vos.NewPercentage(0.15))
	bugetConforto := entities.NewBudgetItem(budget, "Conforto", vos.NewPercentage(0.10))
	budgetPrazeres := entities.NewBudgetItem(budget, "Prazeres", vos.NewPercentage(0.10))
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

	budgetLiberdadeFinanceira.AddAmountUsed(vos.NewMoney(2_774.88))

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
	confortoSum := linq.Sum(conforto, func(i dtos.BillItemQuery) float64 { return i.Total })
	bugetConforto.AddAmountUsed(vos.NewMoney(confortoSum))

	custoFixo := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return !strings.Contains(b.Description, "Faxina")
	})
	custoFixoSum := linq.Sum(custoFixo, func(i dtos.BillItemQuery) float64 { return i.Total })
	budgetCustoFixos.AddAmountUsed(vos.NewMoney(custoFixoSum))

	// gastos de cartão por categoria
	cardByCategory := map[string]float64{}
	groupedByTag := linq.GroupBy(invoices.Items, func(i dtos.InvoiceItemRead) string {
		return i.Tags
	})
	for tag, group := range groupedByTag {
		sum := linq.Sum(group, func(i dtos.InvoiceItemRead) float64 { return i.InstallmentValue })
		cardByCategory[tag] = sum
		for _, item := range budget.Items {
			if item.Category == tag {
				item.AddAmountUsed(vos.NewMoney(sum))
			}
		}
	}

	// gastos de outros (boletos/contas) por categoria
	othersByCategory := map[string]float64{
		"Conforto":    confortoSum,
		"Custos fixos": custoFixoSum,
		"Liberdade Financeira": 2_774.88,
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{
		"Mês", "Categoria", "%", "Devo Gastar",
		"Gasto Cartão", "Gasto Outros", "Total Gasto", "Ainda Posso Gastar",
	})

	for _, item := range budget.Items {
		spentCard := cardByCategory[item.Category]
		spentOthers := othersByCategory[item.Category]
		totalSpent := spentCard + spentOthers
		remaining := item.AmountGoal.Sub(vos.NewMoney(totalSpent))

		table.Append([]string{
			budget.Date.Format("January 2006"),
			item.Category,
			formatPercentage(item.PercentageGoal),
			formatMoney(item.AmountGoal),
			formatMoney(vos.NewMoney(spentCard)),
			formatMoney(vos.NewMoney(spentOthers)),
			formatMoney(vos.NewMoney(totalSpent)),
			formatMoney(remaining),
		})
	}

	table.Render()
}

func RunBudgetFullView(dateParam time.Time) {
	environments.SetupEnvironments()
	db := database.NewConnection()
	ioc.SetupDependencyInjection(db)

	date := time.Date(dateParam.Year(), dateParam.Month(), 1, 0, 0, 0, 0, time.UTC)

	budget := entities.NewBudget(date, vos.NewMoney(13_874.40))
	budgetMetas := entities.NewBudgetItem(budget, "Metas", vos.NewPercentage(0.15))
	bugetConforto := entities.NewBudgetItem(budget, "Conforto", vos.NewPercentage(0.10))
	budgetPrazeres := entities.NewBudgetItem(budget, "Prazeres", vos.NewPercentage(0.10))
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

	budgetLiberdadeFinanceira.AddAmountUsed(vos.NewMoney(2_774.88))

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
	confortoSum := linq.Sum(conforto, func(i dtos.BillItemQuery) float64 { return i.Total })
	bugetConforto.AddAmountUsed(vos.NewMoney(confortoSum))

	custoFixo := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return !strings.Contains(b.Description, "Faxina")
	})
	custoFixoSum := linq.Sum(custoFixo, func(i dtos.BillItemQuery) float64 { return i.Total })
	budgetCustoFixos.AddAmountUsed(vos.NewMoney(custoFixoSum))

	groupedByTag := linq.GroupBy(invoices.Items, func(i dtos.InvoiceItemRead) string {
		return i.Tags
	})

	for tag, group := range groupedByTag {
		sum := linq.Sum(group, func(i dtos.InvoiceItemRead) float64 {
			return i.InstallmentValue
		})
		for _, item := range budget.Items {
			if item.Category == tag {
				item.AddAmountUsed(vos.NewMoney(sum))
			}
		}
	}

	// Tabela 1: visão do orçamento
	fmt.Println("\n=== Orçamento ===")
	budgetData := [][]string{}
	budgetTable := tablewriter.NewWriter(os.Stdout)
	budgetTable.Header([]string{"Mês", "Total a Gastar (Mensal)", "Orçamento", "Orçamento %", "Devo Gastar", "Valor Gasto", "Ainda Posso Gastar"})
	for _, item := range budget.Items {
		budgetData = append(budgetData, []string{
			budget.Date.Format("January 2006"),
			formatMoney(budget.AmountGoal),
			item.Category,
			formatPercentage(item.PercentageGoal),
			formatMoney(item.AmountGoal),
			formatMoney(item.AmountUsed),
			formatMoney(item.AmountGoal.Sub(item.AmountUsed)),
		})
	}
	for _, v := range budgetData {
		budgetTable.Append(v)
	}
	budgetTable.Render()

	// Tabela 2: visão cartão x outros
	gastoPrazeres := &BudgetCardAndOthers{Date: date, Category: "Prazeres", TotalSpentOthers: vos.NewMoney(0)}
	gastoConhecimento := &BudgetCardAndOthers{Date: date, Category: "Conhecimento", TotalSpentOthers: vos.NewMoney(0)}
	gastoConforto := &BudgetCardAndOthers{Date: date, Category: "Conforto", TotalSpentOthers: vos.NewMoney(confortoSum)}
	gastoCustoFixos := &BudgetCardAndOthers{Date: date, Category: "Custos fixos", TotalSpentOthers: vos.NewMoney(custoFixoSum)}

	budgetCard := &BudgetCard{
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
		for _, item := range budgetCard.Items {
			if item.Category == tag {
				item.AddAmountUsed(vos.NewMoney(sum))
			}
		}
	}

	fmt.Println("\n=== Gastos por Cartão e Outros ===")
	cardData := [][]string{}
	cardTable := tablewriter.NewWriter(os.Stdout)
	cardTable.Header([]string{"Mês", "Categoria", "Total Gasto (Cartão)", "Total Gasto (Outros)", "Total"})
	for _, item := range budgetCard.Items {
		cardData = append(cardData, []string{
			item.Date.Format("January 2006"),
			item.Category,
			formatMoney(item.TotalSpentCard),
			formatMoney(item.TotalSpentOthers),
			formatMoney(item.Total),
		})
	}
	for _, v := range cardData {
		cardTable.Append(v)
	}
	cardTable.Render()
}

func RunBalance(dateParam time.Time) {
	environments.SetupEnvironments()
	db := database.NewConnection()
	ioc.SetupDependencyInjection(db)

	date := time.Date(dateParam.Year(), dateParam.Month(), 1, 0, 0, 0, 0, time.UTC)

	budget := entities.NewBudget(date, vos.NewMoney(13_874.40))
	budgetMetas := entities.NewBudgetItem(budget, "Metas", vos.NewPercentage(0.15))
	bugetConforto := entities.NewBudgetItem(budget, "Conforto", vos.NewPercentage(0.10))
	budgetPrazeres := entities.NewBudgetItem(budget, "Prazeres", vos.NewPercentage(0.10))
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

	budgetLiberdadeFinanceira.AddAmountUsed(vos.NewMoney(2_774.88))

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
	confortoSum := linq.Sum(conforto, func(i dtos.BillItemQuery) float64 { return i.Total })
	bugetConforto.AddAmountUsed(vos.NewMoney(confortoSum))

	custoFixo := linq.Filter(bills.Items, func(b dtos.BillItemQuery) bool {
		return !strings.Contains(b.Description, "Faxina")
	})
	custoFixoSum := linq.Sum(custoFixo, func(i dtos.BillItemQuery) float64 { return i.Total })
	budgetCustoFixos.AddAmountUsed(vos.NewMoney(custoFixoSum))

	knownCategories := map[string]bool{}
	for _, item := range budget.Items {
		knownCategories[item.Category] = true
	}

	groupedByTag := linq.GroupBy(invoices.Items, func(i dtos.InvoiceItemRead) string {
		return i.Tags
	})

	cardByCategory := map[string][]dtos.InvoiceItemRead{}
	var othersSum float64

	for tag, group := range groupedByTag {
		if !knownCategories[tag] {
			sum := linq.Sum(group, func(i dtos.InvoiceItemRead) float64 { return i.InstallmentValue })
			othersSum += sum
			continue
		}
		cardByCategory[tag] = group
		sum := linq.Sum(group, func(i dtos.InvoiceItemRead) float64 { return i.InstallmentValue })
		for _, item := range budget.Items {
			if item.Category == tag {
				item.AddAmountUsed(vos.NewMoney(sum))
			}
		}
	}

	if othersSum > 0 {
		budgetCustoFixos.AddAmountUsed(vos.NewMoney(othersSum))
	}

	fmt.Println("\n=== Saldo Disponível Real ===")
	balanceData := [][]string{}
	balanceTable := tablewriter.NewWriter(os.Stdout)
	balanceTable.Header([]string{"Mês", "Total a Gastar (Mensal)", "Categoria", "Planejado", "Valor Gasto", "Saldo Restante"})

	for _, item := range budget.Items {
		balanceData = append(balanceData, []string{
			budget.Date.Format("January 2006"),
			formatMoney(budget.AmountGoal),
			item.Category,
			formatMoney(item.AmountGoal),
			formatMoney(item.AmountUsed),
			formatMoney(item.AmountGoal.Sub(item.AmountUsed)),
		})
	}

	for _, v := range balanceData {
		balanceTable.Append(v)
	}
	balanceTable.Render()

	fmt.Println("\n=== Cartão de Crédito — Saldo por Subcategoria ===")
	cardData := [][]string{}
	cardTable := tablewriter.NewWriter(os.Stdout)
	cardTable.Header([]string{"Mês", "Categoria", "Subcategoria", "Devo Gastar", "Valor Gasto", "Saldo Restante"})

	for _, budgetItem := range budget.Items {
		items, ok := cardByCategory[budgetItem.Category]
		if !ok {
			continue
		}

		tagBudget := budgetItem.AmountGoal.Money()
		tagCardTotal := linq.Sum(items, func(i dtos.InvoiceItemRead) float64 { return i.InstallmentValue })
		tagRemaining := tagBudget - tagCardTotal

		subGrouped := linq.GroupBy(items, func(i dtos.InvoiceItemRead) string {
			if i.Category.Name == "" {
				return "Sem subcategoria"
			}
			return i.Category.Name
		})

		for subcat, subItems := range subGrouped {
			subSum := linq.Sum(subItems, func(i dtos.InvoiceItemRead) float64 { return i.InstallmentValue })
			cardData = append(cardData, []string{
				budget.Date.Format("January 2006"),
				budgetItem.Category,
				subcat,
				formatMoney(budgetItem.AmountGoal),
				formatMoney(vos.NewMoney(subSum)),
				formatMoney(vos.NewMoney(tagRemaining)),
			})
		}
	}

	for _, v := range cardData {
		cardTable.Append(v)
	}
	cardTable.Render()
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

func RunBudgetByCategory(dateParam time.Time, category string) {
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

	fmt.Printf("\n=== Itens da Categoria: %s ===\n", category)
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Data da Compra", "Descrição", "Valor", "Origem"})

	var total float64

	// Itens do cartão (invoices) agrupados por tag
	groupedByTag := linq.GroupBy(invoices.Items, func(i dtos.InvoiceItemRead) string {
		return i.Tags
	})

	for tag, group := range groupedByTag {
		if !strings.EqualFold(tag, category) {
			continue
		}
		for _, item := range group {
			table.Append([]string{
				item.PurchaseDate.Format("02/01/2006"),
				item.Description,
				formatMoney(vos.NewMoney(item.InstallmentValue)),
				"Cartão",
			})
			total += item.InstallmentValue
		}
	}

	// Itens de boletos/contas (bills) filtrados por categoria
	if strings.EqualFold(category, "Conforto") {
		for _, item := range bills.Items {
			if strings.Contains(item.Description, "Faxina") {
				table.Append([]string{
					date.Format("02/01/2006"),
					item.Description,
					formatMoney(vos.NewMoney(item.Total)),
					"Boleto/Conta",
				})
				total += item.Total
			}
		}
	} else if strings.EqualFold(category, "Custos fixos") {
		for _, item := range bills.Items {
			if !strings.Contains(item.Description, "Faxina") {
				table.Append([]string{
					date.Format("02/01/2006"),
					item.Description,
					formatMoney(vos.NewMoney(item.Total)),
					"Boleto/Conta",
				})
				total += item.Total
			}
		}
	}

	table.Render()
	fmt.Printf("\nTotal: %s\n", formatMoney(vos.NewMoney(total)))
}

func formatMoney(value vos.Money) string {
	format := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ".", Decimal: ","}
	return format.FormatMoney(value.Money())
}

func formatPercentage(value vos.Percentage) string {
	return fmt.Sprintf("%.2f%%", value.Percentage()*100)
}
