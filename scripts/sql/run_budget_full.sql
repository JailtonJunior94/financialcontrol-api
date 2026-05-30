-- run_budget_full.sql
-- Replica: make run_budget_full DATE=DD/MM/YYYY
-- Gera dois resultsets em sequencia:
--   1. === Orcamento ===            (identico a run_budget)
--   2. === Gastos por Cartao e Outros === (identico a run_budget_cards_and_others)
-- Parametros: $(date) = primeiro dia do mes de referencia (ex: 2026-06-01)
--             $(invoiceDate) = date + 1 mes (ex: 2026-07-01)

DECLARE @refDate DATETIME = CONVERT(DATETIME, '$(date)');
DECLARE @invDate DATETIME = CONVERT(DATETIME, '$(invoiceDate)');
DECLARE @totalMensal DECIMAL(18,2) = 13874.40;
DECLARE @liberdadeUsada DECIMAL(18,2) = 2774.88;

-- ========================================================
-- Resultset 1: Orcamento
-- ========================================================
PRINT '=== Orcamento ===';

WITH Budget AS (
    SELECT categoria, percentual, ROUND(@totalMensal * percentual / 100.0, 2) AS devo_gastar
    FROM (VALUES
        ('Custos fixos',         40.0),
        ('Conforto',             10.0),
        ('Metas',                15.0),
        ('Prazeres',             10.0),
        ('Conhecimento',          5.0),
        ('Liberdade Financeira', 20.0)
    ) AS t(categoria, percentual)
),
Bills AS (
    SELECT
        CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END AS categoria,
        SUM(bi.Value) AS total
    FROM dbo.Bill b (NOLOCK)
    INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
    WHERE b.[Date] = @refDate AND b.Active = 1
    GROUP BY CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END
),
Invoices AS (
    SELECT ii.Tags AS categoria, SUM(ii.InstallmentValue) AS total
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    WHERE i.[Date] = @invDate AND ii.Tags != ''
    GROUP BY ii.Tags
),
GastoTotal AS (
    SELECT categoria, SUM(total) AS total FROM (
        SELECT categoria, total FROM Bills
        UNION ALL SELECT categoria, total FROM Invoices
    ) AS x GROUP BY categoria
),
LiberdadeFixa AS (SELECT 'Liberdade Financeira' AS categoria, @liberdadeUsada AS total)
SELECT
    FORMAT(@refDate, 'MMMM yyyy', 'pt-BR')      AS [Mes],
    'R$ ' + FORMAT(@totalMensal, 'N2', 'pt-BR') AS [Total a Gastar (Mensal)],
    b.categoria                                  AS [Orcamento],
    FORMAT(b.percentual, 'N2') + '%'             AS [Orcamento %],
    'R$ ' + FORMAT(b.devo_gastar, 'N2', 'pt-BR') AS [Devo Gastar],
    'R$ ' + FORMAT(ISNULL(
        CASE b.categoria WHEN 'Liberdade Financeira' THEN lf.total ELSE g.total END, 0),
        'N2', 'pt-BR')                           AS [Valor Gasto],
    'R$ ' + FORMAT(b.devo_gastar - ISNULL(
        CASE b.categoria WHEN 'Liberdade Financeira' THEN lf.total ELSE g.total END, 0),
        'N2', 'pt-BR')                           AS [Ainda Posso Gastar]
FROM Budget b
LEFT JOIN GastoTotal g ON g.categoria = b.categoria
LEFT JOIN LiberdadeFixa lf ON lf.categoria = b.categoria
ORDER BY CASE b.categoria
    WHEN 'Custos fixos' THEN 1 WHEN 'Conforto' THEN 2 WHEN 'Metas' THEN 3
    WHEN 'Prazeres' THEN 4 WHEN 'Conhecimento' THEN 5 WHEN 'Liberdade Financeira' THEN 6 END;

-- ========================================================
-- Resultset 2: Gastos por Cartao e Outros
-- ========================================================
PRINT '=== Gastos por Cartao e Outros ===';

WITH CardSpend AS (
    SELECT ii.Tags AS categoria, SUM(ii.InstallmentValue) AS total_cartao
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    WHERE i.[Date] = @invDate AND ii.Tags != ''
    GROUP BY ii.Tags
),
OthersSpend AS (
    SELECT
        CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END AS categoria,
        SUM(bi.Value) AS total_outros
    FROM dbo.Bill b (NOLOCK)
    INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
    WHERE b.[Date] = @refDate AND b.Active = 1
    GROUP BY CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END
),
Categories AS (
    SELECT categoria FROM (VALUES
        ('Conforto'), ('Prazeres'), ('Custos fixos'), ('Conhecimento')
    ) AS t(categoria)
)
SELECT
    FORMAT(@refDate, 'MMMM yyyy', 'pt-BR')                      AS [Mes],
    c.categoria                                                   AS [Categoria],
    'R$ ' + FORMAT(ISNULL(cs.total_cartao, 0), 'N2', 'pt-BR')  AS [Total Gasto (Cartao)],
    'R$ ' + FORMAT(ISNULL(os.total_outros, 0), 'N2', 'pt-BR')  AS [Total Gasto (Outros)],
    'R$ ' + FORMAT(ISNULL(cs.total_cartao, 0) + ISNULL(os.total_outros, 0), 'N2', 'pt-BR') AS [Total]
FROM Categories c
LEFT JOIN CardSpend   cs ON cs.categoria = c.categoria
LEFT JOIN OthersSpend os ON os.categoria = c.categoria
ORDER BY CASE c.categoria
    WHEN 'Conforto' THEN 1 WHEN 'Prazeres' THEN 2
    WHEN 'Custos fixos' THEN 3 WHEN 'Conhecimento' THEN 4 END;
GO
