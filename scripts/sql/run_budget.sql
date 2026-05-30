-- run_budget.sql
-- Replica: make run_budget DATE=DD/MM/YYYY
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
Bills AS (
    SELECT
        CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END AS categoria,
        SUM(bi.Value) AS total
    FROM dbo.Bill b (NOLOCK)
    INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
    WHERE b.[Date] = @refDate
      AND b.Active = 1
    GROUP BY CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END
),
Invoices AS (
    SELECT
        ii.Tags AS categoria,
        SUM(ii.InstallmentValue) AS total
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    WHERE i.[Date] = @invDate
      AND ii.Tags != ''
    GROUP BY ii.Tags
),
GastoTotal AS (
    SELECT categoria, SUM(total) AS total
    FROM (
        SELECT categoria, total FROM Bills
        UNION ALL
        SELECT categoria, total FROM Invoices
    ) AS combined
    GROUP BY categoria
),
LiberdadeFixa AS (
    SELECT 'Liberdade Financeira' AS categoria, @liberdadeUsada AS total
)
SELECT
    FORMAT(@refDate, 'MMMM yyyy', 'pt-BR')     AS [Mes],
    '$' + FORMAT(@totalMensal, 'N2', 'pt-BR')  AS [Total a Gastar (Mensal)],
    b.categoria                                 AS [Orcamento],
    FORMAT(b.percentual, 'N2') + '%'            AS [Orcamento %],
    'R$ ' + FORMAT(b.devo_gastar, 'N2', 'pt-BR') AS [Devo Gastar],
    'R$ ' + FORMAT(ISNULL(
        CASE b.categoria
            WHEN 'Liberdade Financeira' THEN lf.total
            ELSE g.total
        END, 0), 'N2', 'pt-BR')                AS [Valor Gasto],
    'R$ ' + FORMAT(b.devo_gastar - ISNULL(
        CASE b.categoria
            WHEN 'Liberdade Financeira' THEN lf.total
            ELSE g.total
        END, 0), 'N2', 'pt-BR')               AS [Ainda Posso Gastar]
FROM Budget b
LEFT JOIN GastoTotal g ON g.categoria = b.categoria
LEFT JOIN LiberdadeFixa lf ON lf.categoria = b.categoria
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
