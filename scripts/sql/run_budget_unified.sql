-- run_budget_unified.sql
-- Replica: make run_budget_unified DATE=DD/MM/YYYY
-- Parametros: $(date) = primeiro dia do mes de referencia (ex: 2026-06-01)
--             $(invoiceDate) = date + 1 mes (ex: 2026-07-01)

DECLARE @refDate DATETIME = CONVERT(DATETIME, '$(date)');
DECLARE @invDate DATETIME = CONVERT(DATETIME, '$(invoiceDate)');
DECLARE @totalMensal DECIMAL(18,2) = 13874.40;
DECLARE @liberdadeUsada DECIMAL(18,2) = 2774.88;

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
CardSpend AS (
    SELECT ii.Tags AS categoria, SUM(ii.InstallmentValue) AS total
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    WHERE i.[Date] = @invDate AND ii.Tags != ''
    GROUP BY ii.Tags
),
OthersSpend AS (
    SELECT
        CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END AS categoria,
        SUM(bi.Value) AS total
    FROM dbo.Bill b (NOLOCK)
    INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
    WHERE b.[Date] = @refDate AND b.Active = 1
    GROUP BY CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END
),
OthersFixed AS (
    SELECT 'Liberdade Financeira' AS categoria, @liberdadeUsada AS total
)
SELECT
    FORMAT(@refDate, 'MMMM yyyy', 'pt-BR')                          AS [Mes],
    b.categoria                                                       AS [Categoria],
    FORMAT(b.percentual, 'N2') + '%'                                 AS [%],
    'R$ ' + FORMAT(b.devo_gastar, 'N2', 'pt-BR')                    AS [Devo Gastar],
    'R$ ' + FORMAT(ISNULL(cs.total, 0), 'N2', 'pt-BR')             AS [Gasto Cartao],
    'R$ ' + FORMAT(
        ISNULL(CASE b.categoria WHEN 'Liberdade Financeira' THEN of2.total ELSE os.total END, 0),
        'N2', 'pt-BR')                                               AS [Gasto Outros],
    'R$ ' + FORMAT(
        ISNULL(cs.total, 0) +
        ISNULL(CASE b.categoria WHEN 'Liberdade Financeira' THEN of2.total ELSE os.total END, 0),
        'N2', 'pt-BR')                                               AS [Total Gasto],
    'R$ ' + FORMAT(
        b.devo_gastar -
        ISNULL(cs.total, 0) -
        ISNULL(CASE b.categoria WHEN 'Liberdade Financeira' THEN of2.total ELSE os.total END, 0),
        'N2', 'pt-BR')                                               AS [Ainda Posso Gastar]
FROM Budget b
LEFT JOIN CardSpend  cs  ON cs.categoria  = b.categoria
LEFT JOIN OthersSpend os  ON os.categoria  = b.categoria
LEFT JOIN OthersFixed of2 ON of2.categoria = b.categoria
ORDER BY
    CASE b.categoria
        WHEN 'Custos fixos'         THEN 1
        WHEN 'Conforto'             THEN 2
        WHEN 'Metas'                THEN 3
        WHEN 'Prazeres'             THEN 4
        WHEN 'Conhecimento'         THEN 5
        WHEN 'Liberdade Financeira' THEN 6
    END;
GO
