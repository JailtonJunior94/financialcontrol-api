package application

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning"

	"github.com/JailtonJunior94/devkit-go/pkg/linq"
	"github.com/JailtonJunior94/devkit-go/pkg/vos"
	"github.com/leekchan/accounting"
	"github.com/olekukonko/tablewriter"
)

// PlanningRunner implements all CLI planning operations through explicit ports.
// It replaces the direct container access that previously existed in bootstrap/cli.
type PlanningRunner struct {
	billing   planning.BillingReadPort
	invoicing planning.InvoicingReadPort
	sync      planning.SyncPort
}

func NewPlanningRunner(
	billing planning.BillingReadPort,
	invoicing planning.InvoicingReadPort,
	sync planning.SyncPort,
) *PlanningRunner {
	return &PlanningRunner{
		billing:   billing,
		invoicing: invoicing,
		sync:      sync,
	}
}

func (r *PlanningRunner) RunBudget(date time.Time) error {
	ctx := context.Background()
	ref := firstDayOf(date)

	bills, err := r.billing.GetMonthlyBills(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching bills: %w", err)
	}

	invoices, err := r.invoicing.GetMonthlyInvoices(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching invoices: %w", err)
	}

	budget, items := newDefaultBudget(ref)
	applyBillsToBudget(bills, items)
	applyInvoicesToBudget(invoices, items)

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Mês", "Total a Gastar (Mensal)", "Orçamento", "Orçamento %", "Devo Gastar", "Valor Gasto", "Ainda Posso Gastar"})
	for _, item := range budget.Items {
		remaining, _ := item.AmountGoal.Subtract(item.AmountUsed)
		if err := table.Append([]string{
			budget.Date.Format("January 2006"),
			formatMoney(budget.AmountGoal),
			item.Category,
			formatPercentage(item.PercentageGoal),
			formatMoney(item.AmountGoal),
			formatMoney(item.AmountUsed),
			formatMoney(remaining),
		}); err != nil {
			return fmt.Errorf("planning: appending budget row: %w", err)
		}
	}
	if err := table.Render(); err != nil {
		return fmt.Errorf("planning: rendering budget table: %w", err)
	}
	return nil
}

func (r *PlanningRunner) RunBudgetCardsAndOthers(date time.Time) error {
	ctx := context.Background()
	ref := firstDayOf(date)

	bills, err := r.billing.GetMonthlyBills(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching bills: %w", err)
	}

	invoices, err := r.invoicing.GetMonthlyInvoices(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching invoices: %w", err)
	}

	confortoSum := sumBills(bills.Items, "Faxina")
	custoFixoSum := sumBillsExcluding(bills.Items, "Faxina")

	budgetCard := &budgetCardView{
		Items: []*budgetCardAndOthers{
			{Date: ref, Category: "Conforto", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(confortoSum)},
			{Date: ref, Category: "Prazeres", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(0)},
			{Date: ref, Category: "Custos fixos", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(custoFixoSum)},
			{Date: ref, Category: "Conhecimento", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(0)},
		},
	}
	applyInvoicesToCardView(invoices, budgetCard)

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Mês", "Categoria", "Total Gasto (Cartão)", "Total Gasto (Outros)", "Total"})
	for _, item := range budgetCard.Items {
		if err := table.Append([]string{
			item.Date.Format("January 2006"),
			item.Category,
			formatMoney(item.TotalSpentCard),
			formatMoney(item.TotalSpentOthers),
			formatMoney(item.Total),
		}); err != nil {
			return fmt.Errorf("planning: appending budget cards row: %w", err)
		}
	}
	if err := table.Render(); err != nil {
		return fmt.Errorf("planning: rendering budget cards table: %w", err)
	}
	return nil
}

func (r *PlanningRunner) RunBudgetUnified(date time.Time) error {
	ctx := context.Background()
	ref := firstDayOf(date)

	bills, err := r.billing.GetMonthlyBills(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching bills: %w", err)
	}

	invoices, err := r.invoicing.GetMonthlyInvoices(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching invoices: %w", err)
	}

	budget, items := newDefaultBudget(ref)
	applyBillsToBudget(bills, items)
	applyInvoicesToBudget(invoices, items)

	confortoSum := sumBills(bills.Items, "Faxina")
	custoFixoSum := sumBillsExcluding(bills.Items, "Faxina")

	cardByCategory := map[string]float64{}
	groupedByTag := linq.GroupBy(invoices.Items, func(i planning.InvoiceReadItem) string { return i.Tags })
	for tag, group := range groupedByTag {
		cardByCategory[tag] = linq.Sum(group, func(i planning.InvoiceReadItem) float64 { return i.Total })
	}

	othersByCategory := map[string]float64{
		"Conforto":             confortoSum,
		"Custos fixos":         custoFixoSum,
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
		spentMoney := mustMoneyBRL(totalSpent)
		remaining, _ := item.AmountGoal.Subtract(spentMoney)

		if err := table.Append([]string{
			budget.Date.Format("January 2006"),
			item.Category,
			formatPercentage(item.PercentageGoal),
			formatMoney(item.AmountGoal),
			formatMoney(mustMoneyBRL(spentCard)),
			formatMoney(mustMoneyBRL(spentOthers)),
			formatMoney(mustMoneyBRL(totalSpent)),
			formatMoney(remaining),
		}); err != nil {
			return fmt.Errorf("planning: appending unified budget row: %w", err)
		}
	}
	if err := table.Render(); err != nil {
		return fmt.Errorf("planning: rendering unified budget table: %w", err)
	}
	return nil
}

func (r *PlanningRunner) RunBudgetFull(date time.Time) error {
	ctx := context.Background()
	ref := firstDayOf(date)

	bills, err := r.billing.GetMonthlyBills(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching bills: %w", err)
	}

	invoices, err := r.invoicing.GetMonthlyInvoices(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching invoices: %w", err)
	}

	budget, items := newDefaultBudget(ref)
	applyBillsToBudget(bills, items)
	applyInvoicesToBudget(invoices, items)

	fmt.Println("\n=== Orçamento ===")
	budgetTable := tablewriter.NewWriter(os.Stdout)
	budgetTable.Header([]string{"Mês", "Total a Gastar (Mensal)", "Orçamento", "Orçamento %", "Devo Gastar", "Valor Gasto", "Ainda Posso Gastar"})
	for _, item := range budget.Items {
		remaining, _ := item.AmountGoal.Subtract(item.AmountUsed)
		if err := budgetTable.Append([]string{
			budget.Date.Format("January 2006"),
			formatMoney(budget.AmountGoal),
			item.Category,
			formatPercentage(item.PercentageGoal),
			formatMoney(item.AmountGoal),
			formatMoney(item.AmountUsed),
			formatMoney(remaining),
		}); err != nil {
			return fmt.Errorf("planning: appending full budget row: %w", err)
		}
	}
	if err := budgetTable.Render(); err != nil {
		return fmt.Errorf("planning: rendering full budget table: %w", err)
	}

	confortoSum := sumBills(bills.Items, "Faxina")
	custoFixoSum := sumBillsExcluding(bills.Items, "Faxina")

	budgetCard := &budgetCardView{
		Items: []*budgetCardAndOthers{
			{Date: ref, Category: "Conforto", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(confortoSum)},
			{Date: ref, Category: "Prazeres", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(0)},
			{Date: ref, Category: "Custos fixos", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(custoFixoSum)},
			{Date: ref, Category: "Conhecimento", TotalSpentCard: mustMoneyBRL(0), TotalSpentOthers: mustMoneyBRL(0)},
		},
	}
	applyInvoicesToCardView(invoices, budgetCard)

	fmt.Println("\n=== Gastos por Cartão e Outros ===")
	cardTable := tablewriter.NewWriter(os.Stdout)
	cardTable.Header([]string{"Mês", "Categoria", "Total Gasto (Cartão)", "Total Gasto (Outros)", "Total"})
	for _, item := range budgetCard.Items {
		if err := cardTable.Append([]string{
			item.Date.Format("January 2006"),
			item.Category,
			formatMoney(item.TotalSpentCard),
			formatMoney(item.TotalSpentOthers),
			formatMoney(item.Total),
		}); err != nil {
			return fmt.Errorf("planning: appending full cards row: %w", err)
		}
	}
	if err := cardTable.Render(); err != nil {
		return fmt.Errorf("planning: rendering full cards table: %w", err)
	}
	return nil
}

func (r *PlanningRunner) RunBalance(date time.Time) error {
	ctx := context.Background()
	ref := firstDayOf(date)

	bills, err := r.billing.GetMonthlyBills(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching bills: %w", err)
	}

	invoices, err := r.invoicing.GetMonthlyInvoices(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching invoices: %w", err)
	}

	budget, items := newDefaultBudget(ref)
	applyBillsToBudget(bills, items)

	knownCategories := map[string]bool{}
	for _, item := range budget.Items {
		knownCategories[item.Category] = true
	}

	cardByCategory := map[string][]planning.InvoiceReadItem{}
	var othersSum float64

	groupedByTag := linq.GroupBy(invoices.Items, func(i planning.InvoiceReadItem) string { return i.Tags })
	for tag, group := range groupedByTag {
		if !knownCategories[tag] {
			othersSum += linq.Sum(group, func(i planning.InvoiceReadItem) float64 { return i.Total })
			continue
		}
		cardByCategory[tag] = group
		sum := linq.Sum(group, func(i planning.InvoiceReadItem) float64 { return i.Total })
		for _, item := range budget.Items {
			if item.Category == tag {
				item.AddAmountUsed(mustMoneyBRL(sum))
			}
		}
	}

	if othersSum > 0 {
		if custoFixos, ok := items["Custos fixos"]; ok {
			custoFixos.AddAmountUsed(mustMoneyBRL(othersSum))
		}
	}

	fmt.Println("\n=== Saldo Disponível Real ===")
	balanceTable := tablewriter.NewWriter(os.Stdout)
	balanceTable.Header([]string{"Mês", "Total a Gastar (Mensal)", "Categoria", "Planejado", "Valor Gasto", "Saldo Restante"})
	for _, item := range budget.Items {
		remaining, _ := item.AmountGoal.Subtract(item.AmountUsed)
		if err := balanceTable.Append([]string{
			budget.Date.Format("January 2006"),
			formatMoney(budget.AmountGoal),
			item.Category,
			formatMoney(item.AmountGoal),
			formatMoney(item.AmountUsed),
			formatMoney(remaining),
		}); err != nil {
			return fmt.Errorf("planning: appending balance row: %w", err)
		}
	}
	if err := balanceTable.Render(); err != nil {
		return fmt.Errorf("planning: rendering balance table: %w", err)
	}

	fmt.Println("\n=== Cartão de Crédito — Saldo por Subcategoria ===")
	cardTable := tablewriter.NewWriter(os.Stdout)
	cardTable.Header([]string{"Mês", "Categoria", "Subcategoria", "Devo Gastar", "Valor Gasto", "Saldo Restante"})

	for _, budgetItem := range budget.Items {
		categoryItems, ok := cardByCategory[budgetItem.Category]
		if !ok {
			continue
		}

		tagCardTotal := linq.Sum(categoryItems, func(i planning.InvoiceReadItem) float64 { return i.Total })
		tagRemaining := budgetItem.AmountGoal.Float() - tagCardTotal

		subGrouped := linq.GroupBy(categoryItems, func(i planning.InvoiceReadItem) string {
			if i.Category == "" {
				return "Sem subcategoria"
			}
			return i.Category
		})

		for subcat, subItems := range subGrouped {
			subSum := linq.Sum(subItems, func(i planning.InvoiceReadItem) float64 { return i.Total })
			if err := cardTable.Append([]string{
				budget.Date.Format("January 2006"),
				budgetItem.Category,
				subcat,
				formatMoney(budgetItem.AmountGoal),
				formatMoney(mustMoneyBRL(subSum)),
				formatMoney(mustMoneyBRL(tagRemaining)),
			}); err != nil {
				return fmt.Errorf("planning: appending balance card row: %w", err)
			}
		}
	}
	if err := cardTable.Render(); err != nil {
		return fmt.Errorf("planning: rendering balance card table: %w", err)
	}
	return nil
}

func (r *PlanningRunner) RunBudgetByCategory(date time.Time, category string) error {
	ctx := context.Background()
	ref := firstDayOf(date)

	bills, err := r.billing.GetMonthlyBills(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching bills: %w", err)
	}

	invoices, err := r.invoicing.GetMonthlyInvoices(ctx, ref)
	if err != nil {
		return fmt.Errorf("planning: fetching invoices: %w", err)
	}

	fmt.Printf("\n=== Itens da Categoria: %s ===\n", category)
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Data da Compra", "Descrição", "Valor", "Origem"})

	var total float64

	groupedByTag := linq.GroupBy(invoices.Items, func(i planning.InvoiceReadItem) string { return i.Tags })
	for tag, group := range groupedByTag {
		if !strings.EqualFold(tag, category) {
			continue
		}
		for _, item := range group {
			if err := table.Append([]string{
				item.Date.Format("02/01/2006"),
				item.Description,
				formatMoney(mustMoneyBRL(item.Total)),
				"Cartão",
			}); err != nil {
				return fmt.Errorf("planning: appending category card row: %w", err)
			}
			total += item.Total
		}
	}

	if strings.EqualFold(category, "Conforto") {
		for _, item := range bills.Items {
			if strings.Contains(item.Description, "Faxina") {
				if err := table.Append([]string{
					ref.Format("02/01/2006"),
					item.Description,
					formatMoney(mustMoneyBRL(item.Total)),
					"Boleto/Conta",
				}); err != nil {
					return fmt.Errorf("planning: appending conforto row: %w", err)
				}
				total += item.Total
			}
		}
	} else if strings.EqualFold(category, "Custos fixos") {
		for _, item := range bills.Items {
			if !strings.Contains(item.Description, "Faxina") {
				if err := table.Append([]string{
					ref.Format("02/01/2006"),
					item.Description,
					formatMoney(mustMoneyBRL(item.Total)),
					"Boleto/Conta",
				}); err != nil {
					return fmt.Errorf("planning: appending custos fixos row: %w", err)
				}
				total += item.Total
			}
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("planning: rendering category table: %w", err)
	}
	fmt.Printf("\nTotal: %s\n", formatMoney(mustMoneyBRL(total)))
	return nil
}

func (r *PlanningRunner) RunSync() error {
	return r.sync.Sync()
}

// --- internal helpers ---

func firstDayOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// mustMoneyBRL creates a BRL Money from a float64 value.
// Errors (e.g. overflow) are silently ignored — only use for display-time helpers.
func mustMoneyBRL(v float64) vos.Money {
	m, _ := vos.NewMoneyFromFloat(v, vos.CurrencyBRL)
	return m
}

// mustPercentage creates a Percentage from a float64 (e.g. 40.0 for 40%).
// Errors are silently ignored — only use for initialisation-time constants.
func mustPercentage(v float64) vos.Percentage {
	p, _ := vos.NewPercentageFromFloat(v)
	return p
}

// newDefaultBudget creates a Budget with the default goals and item map for
// fast lookup. Returns both the Budget (for ordered iteration) and the item
// map (for category lookup by name).
func newDefaultBudget(date time.Time) (*Budget, map[string]*BudgetItem) {
	budget := newBudget(date, mustMoneyBRL(13_874.40))

	custoFixos := newBudgetItem(budget, "Custos fixos", mustPercentage(40.0))
	conforto := newBudgetItem(budget, "Conforto", mustPercentage(10.0))
	metas := newBudgetItem(budget, "Metas", mustPercentage(15.0))
	prazeres := newBudgetItem(budget, "Prazeres", mustPercentage(10.0))
	conhecimento := newBudgetItem(budget, "Conhecimento", mustPercentage(5.0))
	liberdade := newBudgetItem(budget, "Liberdade Financeira", mustPercentage(20.0))
	liberdade.AddAmountUsed(mustMoneyBRL(2_774.88))

	budget.addItems([]*BudgetItem{custoFixos, conforto, metas, prazeres, conhecimento, liberdade})

	items := map[string]*BudgetItem{
		"Custos fixos":         custoFixos,
		"Conforto":             conforto,
		"Metas":                metas,
		"Prazeres":             prazeres,
		"Conhecimento":         conhecimento,
		"Liberdade Financeira": liberdade,
	}
	return budget, items
}

func applyBillsToBudget(bills *planning.MonthlyBillsReadModel, items map[string]*BudgetItem) {
	if bills == nil {
		return
	}
	confortoSum := sumBills(bills.Items, "Faxina")
	custoFixoSum := sumBillsExcluding(bills.Items, "Faxina")
	if v, ok := items["Conforto"]; ok {
		v.AddAmountUsed(mustMoneyBRL(confortoSum))
	}
	if v, ok := items["Custos fixos"]; ok {
		v.AddAmountUsed(mustMoneyBRL(custoFixoSum))
	}
}

func applyInvoicesToBudget(invoices *planning.MonthlyInvoicesReadModel, items map[string]*BudgetItem) {
	if invoices == nil {
		return
	}
	grouped := linq.GroupBy(invoices.Items, func(i planning.InvoiceReadItem) string { return i.Tags })
	for tag, group := range grouped {
		sum := linq.Sum(group, func(i planning.InvoiceReadItem) float64 { return i.Total })
		if v, ok := items[tag]; ok {
			v.AddAmountUsed(mustMoneyBRL(sum))
		}
	}
}

func sumBills(items []planning.BillReadItem, keyword string) float64 {
	filtered := linq.Filter(items, func(b planning.BillReadItem) bool {
		return strings.Contains(b.Description, keyword)
	})
	return linq.Sum(filtered, func(b planning.BillReadItem) float64 { return b.Total })
}

func sumBillsExcluding(items []planning.BillReadItem, keyword string) float64 {
	filtered := linq.Filter(items, func(b planning.BillReadItem) bool {
		return !strings.Contains(b.Description, keyword)
	})
	return linq.Sum(filtered, func(b planning.BillReadItem) float64 { return b.Total })
}

// --- card view types (local to planning application layer) ---

type budgetCardView struct {
	Items []*budgetCardAndOthers
}

type budgetCardAndOthers struct {
	Date             time.Time
	Category         string
	TotalSpentCard   vos.Money
	TotalSpentOthers vos.Money
	Total            vos.Money
}

func (b *budgetCardAndOthers) addAmountUsed(amount vos.Money) {
	b.TotalSpentCard, _ = b.TotalSpentCard.Add(amount)
	b.Total, _ = b.TotalSpentCard.Add(b.TotalSpentOthers)
}

func applyInvoicesToCardView(invoices *planning.MonthlyInvoicesReadModel, view *budgetCardView) {
	if invoices == nil {
		return
	}
	grouped := linq.GroupBy(invoices.Items, func(i planning.InvoiceReadItem) string { return i.Tags })
	for tag, group := range grouped {
		sum := linq.Sum(group, func(i planning.InvoiceReadItem) float64 { return i.Total })
		for _, item := range view.Items {
			if item.Category == tag {
				item.addAmountUsed(mustMoneyBRL(sum))
			}
		}
	}
}

// --- formatting helpers ---

func formatMoney(value vos.Money) string {
	format := accounting.Accounting{Symbol: "R$ ", Precision: 2, Thousand: ".", Decimal: ","}
	return format.FormatMoney(value.Float())
}

func formatPercentage(value vos.Percentage) string {
	return fmt.Sprintf("%.2f%%", value.Float())
}
