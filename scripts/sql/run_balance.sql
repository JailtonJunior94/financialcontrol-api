-- run_balance.sql
-- Replica: make run_balance DATE=DD/MM/YYYY
-- Gera dois resultsets:
--   1. === Saldo Disponivel Real ===
--   2. === Cartao de Credito - Saldo por Subcategoria ===
-- Parametros: $(date) = primeiro dia do mes de referencia (ex: 2026-06-01)
--             $(invoiceDate) = date + 1 mes (ex: 2026-07-01)

DECLARE @refDate DATETIME = CONVERT(DATETIME, '$(date)');
DECLARE @invDate DATETIME = CONVERT(DATETIME, '$(invoiceDate)');
DECLARE @totalMensal DECIMAL(18,2) = 13874.40;
DECLARE @liberdadeUsada DECIMAL(18,2) = 2774.88;

-- Categorias reconhecidas pelo budget
DECLARE @knownCategories TABLE (categoria NVARCHAR(100));
INSERT INTO @knownCategories VALUES
    ('Custos fixos'), ('Conforto'), ('Metas'),
    ('Prazeres'), ('Conhecimento'), ('Liberdade Financeira');

-- ========================================================
-- Resultset 1: Saldo Disponivel Real
-- ========================================================
PRINT '=== Saldo Disponivel Real ===';

WITH Budget AS (
    SELECT categoria, ROUND(@totalMensal * percentual / 100.0, 2) AS devo_gastar
    FROM (VALUES
        ('Custos fixos',         40.0),
        ('Conforto',             10.0),
        ('Metas',                15.0),
        ('Prazeres',             10.0),
        ('Conhecimento',          5.0),
        ('Liberdade Financeira', 20.0)
    ) AS t(categoria, percentual)
),
-- Invoices com tag conhecida → sua categoria; tag desconhecida → Custos fixos
InvoicesGrouped AS (
    SELECT
        CASE WHEN kc.categoria IS NOT NULL THEN ii.Tags ELSE 'Custos fixos' END AS categoria,
        SUM(ii.InstallmentValue) AS total
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    LEFT JOIN @knownCategories kc ON kc.categoria = ii.Tags
    WHERE i.[Date] = @invDate AND ii.Tags != ''
    GROUP BY CASE WHEN kc.categoria IS NOT NULL THEN ii.Tags ELSE 'Custos fixos' END
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
GastoTotal AS (
    SELECT categoria, SUM(total) AS total FROM (
        SELECT categoria, total FROM InvoicesGrouped
        UNION ALL SELECT categoria, total FROM Bills
    ) AS x GROUP BY categoria
),
LiberdadeFixa AS (SELECT 'Liberdade Financeira' AS categoria, @liberdadeUsada AS total)
SELECT
    FORMAT(@refDate, 'MMMM yyyy', 'pt-BR')       AS [Mes],
    'R$ ' + FORMAT(@totalMensal, 'N2', 'pt-BR')  AS [Total a Gastar (Mensal)],
    b.categoria                                   AS [Categoria],
    'R$ ' + FORMAT(b.devo_gastar, 'N2', 'pt-BR') AS [Planejado],
    'R$ ' + FORMAT(ISNULL(
        CASE b.categoria WHEN 'Liberdade Financeira' THEN lf.total ELSE g.total END, 0),
        'N2', 'pt-BR')                            AS [Valor Gasto],
    'R$ ' + FORMAT(b.devo_gastar - ISNULL(
        CASE b.categoria WHEN 'Liberdade Financeira' THEN lf.total ELSE g.total END, 0),
        'N2', 'pt-BR')                            AS [Saldo Restante]
FROM Budget b
LEFT JOIN GastoTotal g ON g.categoria = b.categoria
LEFT JOIN LiberdadeFixa lf ON lf.categoria = b.categoria
ORDER BY CASE b.categoria
    WHEN 'Custos fixos' THEN 1 WHEN 'Conforto' THEN 2 WHEN 'Metas' THEN 3
    WHEN 'Prazeres' THEN 4 WHEN 'Conhecimento' THEN 5 WHEN 'Liberdade Financeira' THEN 6 END;

-- ========================================================
-- Resultset 2: Cartao de Credito - Saldo por Subcategoria
-- ========================================================
PRINT '=== Cartao de Credito - Saldo por Subcategoria ===';

WITH Budget AS (
    SELECT categoria, ROUND(@totalMensal * percentual / 100.0, 2) AS devo_gastar
    FROM (VALUES
        ('Custos fixos',         40.0),
        ('Conforto',             10.0),
        ('Metas',                15.0),
        ('Prazeres',             10.0),
        ('Conhecimento',          5.0),
        ('Liberdade Financeira', 20.0)
    ) AS t(categoria, percentual)
),
InvoicesKnown AS (
    SELECT
        ii.Tags                  AS categoria,
        ISNULL(NULLIF(c2.Name, ''), 'Sem subcategoria') AS subcategoria,
        SUM(ii.InstallmentValue) AS total_subcat,
        SUM(SUM(ii.InstallmentValue)) OVER (PARTITION BY ii.Tags) AS total_categoria
    FROM dbo.Invoice i (NOLOCK)
    INNER JOIN dbo.InvoiceItem ii (NOLOCK) ON ii.InvoiceId = i.Id
    INNER JOIN dbo.Category c2 (NOLOCK) ON c2.Id = ii.CategoryId
    INNER JOIN @knownCategories kc ON kc.categoria = ii.Tags
    WHERE i.[Date] = @invDate AND ii.Tags != ''
    GROUP BY ii.Tags, ISNULL(NULLIF(c2.Name, ''), 'Sem subcategoria')
)
SELECT
    FORMAT(@refDate, 'MMMM yyyy', 'pt-BR')                           AS [Mes],
    inv.categoria                                                      AS [Categoria],
    inv.subcategoria                                                   AS [Subcategoria],
    'R$ ' + FORMAT(b.devo_gastar, 'N2', 'pt-BR')                     AS [Devo Gastar],
    'R$ ' + FORMAT(inv.total_subcat, 'N2', 'pt-BR')                  AS [Valor Gasto],
    'R$ ' + FORMAT(b.devo_gastar - inv.total_categoria, 'N2', 'pt-BR') AS [Saldo Restante]
FROM InvoicesKnown inv
INNER JOIN Budget b ON b.categoria = inv.categoria
ORDER BY inv.categoria, inv.subcategoria;
GO
