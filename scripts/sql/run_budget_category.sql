-- run_budget_category.sql
-- Replica: make run_budget_category DATE=DD/MM/YYYY CATEGORY=<nome>
-- Parametros: $(date)        = primeiro dia do mes de referencia (ex: 2026-06-01)
--             $(invoiceDate) = date + 1 mes (ex: 2026-07-01)
--             $(category)    = nome da categoria (ex: Alimentacao, Conforto, Custos fixos)

DECLARE @refDate  DATETIME     = CONVERT(DATETIME, '$(date)');
DECLARE @invDate  DATETIME     = CONVERT(DATETIME, '$(invoiceDate)');
DECLARE @category NVARCHAR(100) = N'$(category)';

PRINT '=== Itens da Categoria: ' + @category + ' ===';

-- Itens do cartao de credito
SELECT
    FORMAT(ii.PurchaseDate, 'dd/MM/yyyy')             AS [Data da Compra],
    ii.Description                                     AS [Descricao],
    'R$ ' + FORMAT(ii.InstallmentValue, 'N2', 'pt-BR') AS [Valor],
    'Cartao'                                           AS [Origem]
FROM dbo.Invoice i (NOLOCK)
INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
WHERE i.[Date]  = @invDate
  AND ii.Tags   != ''
  AND ii.Tags LIKE '%' + @category + '%'

UNION ALL

-- Boletos (Conforto: apenas Faxina; Custos fixos: tudo exceto Faxina; outros: nenhum boleto)
SELECT
    FORMAT(@refDate, 'dd/MM/yyyy')                    AS [Data da Compra],
    bi.Title                                           AS [Descricao],
    'R$ ' + FORMAT(bi.Value, 'N2', 'pt-BR')           AS [Valor],
    'Boleto/Conta'                                     AS [Origem]
FROM dbo.Bill b (NOLOCK)
INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
WHERE b.[Date]  = @refDate
  AND b.Active  = 1
  AND (
        (@category LIKE '%Conforto%'     AND bi.Title LIKE '%Faxina%')
     OR (@category LIKE '%Custos fixos%' AND bi.Title NOT LIKE '%Faxina%')
  )

ORDER BY [Data da Compra], [Origem];

-- Total
SELECT
    'R$ ' + FORMAT(SUM(gasto), 'N2', 'pt-BR') AS [Total]
FROM (
    SELECT SUM(ii.InstallmentValue) AS gasto
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    WHERE i.[Date] = @invDate AND ii.Tags != '' AND ii.Tags LIKE '%' + @category + '%'

    UNION ALL

    SELECT SUM(bi.Value) AS gasto
    FROM dbo.Bill b (NOLOCK)
    INNER JOIN dbo.BillItem bi (NOLOCK) ON bi.BillId = b.Id
    WHERE b.[Date] = @refDate AND b.Active = 1
      AND (
            (@category LIKE '%Conforto%'     AND bi.Title LIKE '%Faxina%')
         OR (@category LIKE '%Custos fixos%' AND bi.Title NOT LIKE '%Faxina%')
      )
) AS totais;
GO
