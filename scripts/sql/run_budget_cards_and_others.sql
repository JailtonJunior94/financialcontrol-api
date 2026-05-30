-- run_budget_cards_and_others.sql
-- Replica: make run_budget_cards_and_others DATE=DD/MM/YYYY
-- Parametros: $(date) = primeiro dia do mes de referencia (ex: 2026-06-01)
--             $(invoiceDate) = date + 1 mes (ex: 2026-07-01)

DECLARE @refDate DATETIME = CONVERT(DATETIME, '$(date)');
DECLARE @invDate DATETIME = CONVERT(DATETIME, '$(invoiceDate)');

WITH CardSpend AS (
    SELECT
        ii.Tags                  AS categoria,
        SUM(ii.InstallmentValue) AS total_cartao
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    WHERE i.[Date] = @invDate
      AND ii.Tags != ''
    GROUP BY ii.Tags
),
OthersSpend AS (
    SELECT
        CASE WHEN bi.Title LIKE '%Faxina%' THEN 'Conforto' ELSE 'Custos fixos' END AS categoria,
        SUM(bi.Value) AS total_outros
    FROM dbo.Bill b (NOLOCK)
    INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
    WHERE b.[Date] = @refDate
      AND b.Active = 1
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
    'R$ ' + FORMAT(ISNULL(cs.total_cartao, 0), 'N2', 'pt-BR')   AS [Total Gasto (Cartao)],
    'R$ ' + FORMAT(ISNULL(os.total_outros, 0), 'N2', 'pt-BR')   AS [Total Gasto (Outros)],
    'R$ ' + FORMAT(ISNULL(cs.total_cartao, 0) + ISNULL(os.total_outros, 0), 'N2', 'pt-BR') AS [Total]
FROM Categories c
LEFT JOIN CardSpend  cs ON cs.categoria = c.categoria
LEFT JOIN OthersSpend os ON os.categoria = c.categoria
ORDER BY
    CASE c.categoria
        WHEN 'Conforto'     THEN 1
        WHEN 'Prazeres'     THEN 2
        WHEN 'Custos fixos' THEN 3
        WHEN 'Conhecimento' THEN 4
    END;
GO
