-- Migration: create the finance module schema on MSSQL (DDL only — Task 3.0).
-- DML migration data and counts/sums gate will be added in Task 10.0 to this same file.
-- Idempotency is delegated to dbo.schema_migrations; defensive guards removed.
--
-- Fallback (decision A1.c): if CREATE SCHEMA is not permitted in the target environment,
-- prefix table names with "Finance" in dbo (e.g., dbo.FinanceTransactions) and update
-- constants in pkg/database/queries.go accordingly. The finance. prefix is the default.

IF NOT EXISTS (SELECT 1 FROM sys.schemas WHERE name = 'finance')
    EXEC('CREATE SCHEMA finance');

-- ============================================================
-- Table: finance.Transactions
-- ============================================================
CREATE TABLE finance.Transactions (
    [Id]                    UNIQUEIDENTIFIER NOT NULL,
    [UserId]                UNIQUEIDENTIFIER NOT NULL,
    [Description]           NVARCHAR(255)    COLLATE SQL_Latin1_General_CP1_CI_AI NOT NULL, -- C3.a + F2.b (RF-19)
    [Amount]                DECIMAL(19,4)    NOT NULL,
    [Currency]              CHAR(3)          NOT NULL DEFAULT 'BRL',                         -- H2.b (RF-28)
    [OccurredAt]            DATETIME2        NOT NULL,                                        -- UTC
    [TransactionType]       VARCHAR(32)      NOT NULL,
    [PaymentMethod]         VARCHAR(32)      NOT NULL,
    [CardId]                UNIQUEIDENTIFIER NULL,
    [CategoryId]            UNIQUEIDENTIFIER NOT NULL,
    [SubcategoryId]         UNIQUEIDENTIFIER NULL,
    [OriginalTransactionId] UNIQUEIDENTIFIER NULL,
    [LegacyOrigin]          VARCHAR(64)      NULL,   -- F3.a: "<SourceTable>:<uuid>"
    [CreatedAt]             DATETIME2        NOT NULL,
    [UpdatedAt]             DATETIME2        NOT NULL,
    [DeletedAt]             DATETIME2        NULL,
    CONSTRAINT PK_Transactions PRIMARY KEY ([Id]),
    CONSTRAINT FK_Tx_User FOREIGN KEY ([UserId]) REFERENCES dbo.[User]([Id]),
    CONSTRAINT FK_Tx_Original FOREIGN KEY ([OriginalTransactionId]) REFERENCES finance.Transactions([Id]), -- H1.a (self-FK)
    CONSTRAINT CK_Tx_Amount_Positive CHECK ([Amount] > 0),
    CONSTRAINT CK_Tx_Currency_BRL CHECK ([Currency] = 'BRL'),                               -- H2.b: single-currency enforced; relax in future PRD
    CONSTRAINT CK_Tx_Type CHECK ([TransactionType] IN ('income','expense','credit_purchase','installment_purchase','refund')),
    CONSTRAINT CK_Tx_PaymentMethod CHECK ([PaymentMethod] IN ('pix','boleto','ted','debit_card','credit_card','cash'))
);

CREATE INDEX IX_Tx_User_Occurred
    ON finance.Transactions([UserId], [OccurredAt] DESC, [CreatedAt] DESC)
    WHERE [DeletedAt] IS NULL;

CREATE INDEX IX_Tx_User_LegacyOrigin
    ON finance.Transactions([UserId], [LegacyOrigin])
    WHERE [LegacyOrigin] IS NOT NULL;

CREATE UNIQUE INDEX UX_Tx_ActiveRefundPerOriginal
    ON finance.Transactions([OriginalTransactionId])
    WHERE [TransactionType] = 'refund' AND [DeletedAt] IS NULL; -- RF-13 (409 duplicate guard)

-- ============================================================
-- Table: finance.Invoices
-- ============================================================
CREATE TABLE finance.Invoices (
    [Id]           UNIQUEIDENTIFIER NOT NULL,
    [UserId]       UNIQUEIDENTIFIER NOT NULL,
    [CardId]       UNIQUEIDENTIFIER NOT NULL,
    [State]        VARCHAR(16)      NOT NULL,
    [CycleStart]   DATETIME2        NOT NULL,
    [CycleEnd]     DATETIME2        NOT NULL,
    [ClosingDate]  DATETIME2        NOT NULL,
    [DueDate]      DATETIME2        NOT NULL,
    [Total]        DECIMAL(19,4)    NOT NULL DEFAULT 0,
    [Currency]     CHAR(3)          NOT NULL DEFAULT 'BRL', -- H2.b
    [PaidAt]       DATETIME2        NULL,
    [LegacyOrigin] VARCHAR(64)      NULL,
    [CreatedAt]    DATETIME2        NOT NULL,
    [UpdatedAt]    DATETIME2        NOT NULL,
    [DeletedAt]    DATETIME2        NULL,
    CONSTRAINT PK_Invoices PRIMARY KEY ([Id]),
    CONSTRAINT CK_Inv_State CHECK ([State] IN ('open','closed','paid')),
    CONSTRAINT CK_Inv_Currency_BRL CHECK ([Currency] = 'BRL')
);

CREATE INDEX IX_Inv_User_Card_Cycle
    ON finance.Invoices([UserId], [CardId], [CycleEnd] DESC)
    WHERE [DeletedAt] IS NULL;

CREATE INDEX IX_Inv_User_State
    ON finance.Invoices([UserId], [State])
    WHERE [DeletedAt] IS NULL;

-- ============================================================
-- Table: finance.Installments
-- ============================================================
CREATE TABLE finance.Installments (
    [Id]            UNIQUEIDENTIFIER NOT NULL,
    [TransactionId] UNIQUEIDENTIFIER NOT NULL,
    [InvoiceId]     UNIQUEIDENTIFIER NOT NULL,
    [Number]        INT              NOT NULL,
    [Total]         INT              NOT NULL,
    [Amount]        DECIMAL(19,4)    NOT NULL,
    [Currency]      CHAR(3)          NOT NULL DEFAULT 'BRL', -- H2.b
    [Status]        VARCHAR(32)      NOT NULL,
    [LegacyOrigin]  VARCHAR(64)      NULL,
    [CreatedAt]     DATETIME2        NOT NULL,
    [UpdatedAt]     DATETIME2        NOT NULL,
    [DeletedAt]     DATETIME2        NULL,
    CONSTRAINT PK_Installments PRIMARY KEY ([Id]),
    CONSTRAINT FK_Inst_Tx FOREIGN KEY ([TransactionId]) REFERENCES finance.Transactions([Id]),
    CONSTRAINT FK_Inst_Inv FOREIGN KEY ([InvoiceId]) REFERENCES finance.Invoices([Id]),
    CONSTRAINT CK_Inst_Status CHECK ([Status] IN ('scheduled','anticipated','paid_via_invoice','refunded')),
    CONSTRAINT CK_Inst_Currency_BRL CHECK ([Currency] = 'BRL')
);

CREATE INDEX IX_Inst_Invoice
    ON finance.Installments([InvoiceId])
    WHERE [DeletedAt] IS NULL;

CREATE INDEX IX_Inst_Tx
    ON finance.Installments([TransactionId])
    WHERE [DeletedAt] IS NULL;

-- ============================================================
-- Table: finance.IdempotencyKeys  (RF-26)
-- ============================================================
CREATE TABLE finance.IdempotencyKeys (
    [UserId]       UNIQUEIDENTIFIER NOT NULL,
    [Endpoint]     VARCHAR(128)     NOT NULL,
    [Key]          VARCHAR(64)      NOT NULL,
    [RequestHash]  CHAR(64)         NOT NULL, -- sha256 hex
    [ResponseBody] NVARCHAR(MAX)    NOT NULL,
    [StatusCode]   INT              NOT NULL,
    [CreatedAt]    DATETIME2        NOT NULL,
    [ExpiresAt]    DATETIME2        NOT NULL,
    CONSTRAINT PK_Idemp PRIMARY KEY ([UserId], [Endpoint], [Key])
);

CREATE INDEX IX_Idemp_Expires
    ON finance.IdempotencyKeys([ExpiresAt]);

-- ============================================================
-- DML Migration: populate finance.* from legacy dbo tables
-- ============================================================
-- Idempotent: re-execution does not duplicate rows (RF-37, decision F3.a).
-- LegacyOrigin format: '<SourceTable>:<uuid-origem>' (decision F3.a).
-- RF-41: rows with Active=0 are treated as deleted_at IS NOT NULL — skipped.
-- RF-42: recurring entries migrated as regular transactions.
-- RF-35: defaults inferred when source columns are absent.
--
-- Source → Destination mapping:
--   dbo.TransactionItem  (+ parent dbo.[Transaction] for UserId/Date)
--                        → finance.Transactions  (income / expense)
--   dbo.Invoice          → finance.Invoices
--   dbo.InvoiceItem      → finance.Transactions  (credit_purchase / installment_purchase)
--   dbo.InvoiceItem (Installment>1)
--                        → finance.Installments  (one row per installment number)

-- ============================================================
-- Audit: record migration start (idempotent)
-- ============================================================
DECLARE @auditId UNIQUEIDENTIFIER;
IF EXISTS (SELECT 1 FROM finance.MigrationAudit WHERE [Status] IN ('started', 'dml_ok'))
    SELECT TOP 1 @auditId = [Id]
    FROM finance.MigrationAudit
    WHERE [Status] IN ('started', 'dml_ok')
    ORDER BY [StartedAt] DESC;
ELSE
BEGIN
    SET @auditId = NEWID();
    INSERT INTO finance.MigrationAudit
        ([Id], [StartedAt], [FinishedAt], [SourceTable], [RowsRead], [RowsWritten], [Status])
    VALUES
        (@auditId, GETUTCDATE(), NULL, '*', 0, 0, 'started');
END;

-- ============================================================
-- RF-36: Hard validations — abort with RAISERROR on invalid rows
-- ============================================================
DECLARE @badId UNIQUEIDENTIFIER;

-- TransactionItem: value <= 0
IF EXISTS (SELECT 1 FROM dbo.TransactionItem WHERE Active = 1 AND Value <= 0)
BEGIN
    SELECT TOP 1 @badId = Id FROM dbo.TransactionItem WHERE Active = 1 AND Value <= 0;
    RAISERROR('RF-36: TransactionItem.Value <= 0, source id: %s', 16, 1, CAST(@badId AS VARCHAR(36)));
END;

-- TransactionItem: Type not in (INCOME, OUTCOME)
IF EXISTS (
    SELECT 1 FROM dbo.TransactionItem
    WHERE Active = 1 AND UPPER([Type]) NOT IN ('INCOME', 'OUTCOME')
)
BEGIN
    SELECT TOP 1 @badId = Id
    FROM dbo.TransactionItem
    WHERE Active = 1 AND UPPER([Type]) NOT IN ('INCOME', 'OUTCOME');
    RAISERROR('RF-36: TransactionItem.Type invalid, source id: %s', 16, 1, CAST(@badId AS VARCHAR(36)));
END;

-- InvoiceItem: TotalAmount <= 0 or NULL
IF EXISTS (
    SELECT 1 FROM dbo.InvoiceItem WHERE Active = 1 AND (TotalAmount IS NULL OR TotalAmount <= 0)
)
BEGIN
    SELECT TOP 1 @badId = Id
    FROM dbo.InvoiceItem WHERE Active = 1 AND (TotalAmount IS NULL OR TotalAmount <= 0);
    RAISERROR('RF-36: InvoiceItem.TotalAmount <= 0, source id: %s', 16, 1, CAST(@badId AS VARCHAR(36)));
END;

-- InvoiceItem: Description > 255 chars (RF-36)
IF EXISTS (SELECT 1 FROM dbo.InvoiceItem WHERE Active = 1 AND LEN([Description]) > 255)
BEGIN
    SELECT TOP 1 @badId = Id
    FROM dbo.InvoiceItem WHERE Active = 1 AND LEN([Description]) > 255;
    RAISERROR('RF-36: InvoiceItem.Description > 255 chars, source id: %s', 16, 1, CAST(@badId AS VARCHAR(36)));
END;

-- InvoiceItem: NULL CategoryId
IF EXISTS (SELECT 1 FROM dbo.InvoiceItem WHERE Active = 1 AND CategoryId IS NULL)
BEGIN
    SELECT TOP 1 @badId = Id FROM dbo.InvoiceItem WHERE Active = 1 AND CategoryId IS NULL;
    RAISERROR('RF-36: InvoiceItem.CategoryId is NULL, source id: %s', 16, 1, CAST(@badId AS VARCHAR(36)));
END;

-- ============================================================
-- ID-mapping temp tables (required for installment linking)
-- Reuse existing finance IDs on re-execution (idempotent).
-- ============================================================

-- #InvoiceMap: legacy Invoice.Id → new finance.Invoice.Id
SELECT
    inv.[Id]                                                        AS LegacyInvId,
    ISNULL(fi.[Id], NEWID())                                        AS NewInvId,
    c.[UserId]                                                      AS UserId
INTO #InvoiceMap
FROM dbo.Invoice inv
INNER JOIN dbo.Card      c  ON c.[Id]  = inv.[CardId]
LEFT  JOIN finance.Invoices fi
           ON fi.[LegacyOrigin] = 'Invoice:' + CAST(inv.[Id] AS VARCHAR(36))
WHERE inv.[Active] = 1
  AND c.[Active]   = 1;

-- #InvoiceItemMap: legacy InvoiceItem.Id → new finance.Transaction.Id
SELECT
    ii.[Id]                                                         AS LegacyItemId,
    ISNULL(ft.[Id], NEWID())                                        AS NewTxId,
    im.[NewInvId]                                                   AS FinanceInvId,
    im.[UserId]                                                     AS UserId,
    inv.[CardId]                                                    AS CardId,
    ISNULL(ii.[Installment], 1)                                     AS InstallmentCount
INTO #InvoiceItemMap
FROM dbo.InvoiceItem ii
INNER JOIN dbo.Invoice       inv ON inv.[Id] = ii.[InvoiceId]
INNER JOIN #InvoiceMap       im  ON im.[LegacyInvId] = inv.[Id]
LEFT  JOIN finance.Transactions ft
           ON ft.[LegacyOrigin] = 'InvoiceItem:' + CAST(ii.[Id] AS VARCHAR(36))
WHERE ii.[Active]  = 1
  AND inv.[Active] = 1;

-- ============================================================
-- DML 1: finance.Transactions ← dbo.TransactionItem
-- RF-35: payment_method = pix  (no card_id)
-- RF-35: transaction_type = income | expense  (from Type column)
-- CategoryId: placeholder UUID (no FK constraint on this column)
-- ============================================================
INSERT INTO finance.Transactions (
    [Id], [UserId], [Description], [Amount], [Currency], [OccurredAt],
    [TransactionType], [PaymentMethod],
    [CardId], [CategoryId], [SubcategoryId], [OriginalTransactionId],
    [LegacyOrigin], [CreatedAt], [UpdatedAt]
)
SELECT
    NEWID(),
    t.[UserId],
    ISNULL(NULLIF(LTRIM(RTRIM(ti.[Title])), ''), N'Lançamento legado'),
    CAST(ti.[Value] AS DECIMAL(19,4)),
    'BRL',
    t.[Date],
    CASE UPPER(ti.[Type]) WHEN 'INCOME' THEN 'income' ELSE 'expense' END,
    'pix',                                                          -- RF-35: no card → pix
    NULL,
    CONVERT(UNIQUEIDENTIFIER, '00000000-0000-0000-0000-000000000000'), -- placeholder (no FK)
    NULL,
    NULL,
    'TransactionItem:' + CAST(ti.[Id] AS VARCHAR(36)),
    GETUTCDATE(),
    GETUTCDATE()
FROM dbo.TransactionItem ti
INNER JOIN dbo.[Transaction] t ON t.[Id] = ti.[TransactionId]
WHERE ti.[Active] = 1
  AND t.[Active]  = 1
  AND NOT EXISTS (
        SELECT 1 FROM finance.Transactions
        WHERE [LegacyOrigin] = 'TransactionItem:' + CAST(ti.[Id] AS VARCHAR(36))
  );

-- ============================================================
-- DML 2: finance.Invoices ← dbo.Invoice
-- CycleStart/End derived from Invoice.[Date]; ClosingDate from Card.ClosingDay.
-- ============================================================
INSERT INTO finance.Invoices (
    [Id], [UserId], [CardId], [State],
    [CycleStart], [CycleEnd], [ClosingDate], [DueDate],
    [Total], [Currency], [PaidAt],
    [LegacyOrigin], [CreatedAt], [UpdatedAt]
)
SELECT
    im.[NewInvId],
    im.[UserId],
    inv.[CardId],
    'open',
    CAST(DATEFROMPARTS(YEAR(inv.[Date]), MONTH(inv.[Date]), 1) AS DATETIME2),
    CAST(EOMONTH(inv.[Date]) AS DATETIME2),
    CAST(DATEFROMPARTS(YEAR(inv.[Date]), MONTH(inv.[Date]),
         CASE WHEN c.[ClosingDay] BETWEEN 1 AND 28 THEN c.[ClosingDay] ELSE 10 END)
         AS DATETIME2),
    CAST(inv.[Date] AS DATETIME2),
    CAST(ISNULL(inv.[Total], 0) AS DECIMAL(19,4)),
    'BRL',
    NULL,
    'Invoice:' + CAST(inv.[Id] AS VARCHAR(36)),
    GETUTCDATE(),
    GETUTCDATE()
FROM #InvoiceMap im
INNER JOIN dbo.Invoice inv ON inv.[Id] = im.[LegacyInvId]
INNER JOIN dbo.Card    c   ON c.[Id]   = inv.[CardId]
WHERE NOT EXISTS (
        SELECT 1 FROM finance.Invoices
        WHERE [LegacyOrigin] = 'Invoice:' + CAST(inv.[Id] AS VARCHAR(36))
);

-- ============================================================
-- DML 3: finance.Transactions ← dbo.InvoiceItem (credit purchases)
-- RF-35: payment_method = credit_card (card_id not null)
-- RF-35: transaction_type = credit_purchase if installment<=1, installment_purchase if >1
-- ============================================================
INSERT INTO finance.Transactions (
    [Id], [UserId], [Description], [Amount], [Currency], [OccurredAt],
    [TransactionType], [PaymentMethod],
    [CardId], [CategoryId], [SubcategoryId], [OriginalTransactionId],
    [LegacyOrigin], [CreatedAt], [UpdatedAt]
)
SELECT
    im.[NewTxId],
    im.[UserId],
    SUBSTRING(ii.[Description], 1, 255),
    CAST(ii.[TotalAmount] AS DECIMAL(19,4)),
    'BRL',
    ii.[PurchaseDate],
    CASE WHEN im.[InstallmentCount] <= 1 THEN 'credit_purchase' ELSE 'installment_purchase' END,
    'credit_card',                                                  -- RF-35: card_id not null → credit_card
    im.[CardId],
    ii.[CategoryId],
    NULL,
    NULL,
    'InvoiceItem:' + CAST(ii.[Id] AS VARCHAR(36)),
    GETUTCDATE(),
    GETUTCDATE()
FROM #InvoiceItemMap im
INNER JOIN dbo.InvoiceItem ii ON ii.[Id] = im.[LegacyItemId]
WHERE NOT EXISTS (
        SELECT 1 FROM finance.Transactions
        WHERE [LegacyOrigin] = 'InvoiceItem:' + CAST(ii.[Id] AS VARCHAR(36))
);

-- ============================================================
-- DML 4: finance.Installments ← dbo.InvoiceItem (installment_purchase rows only)
-- One finance.Installment per installment number (1..N).
-- ============================================================
WITH Numbers(n) AS (
    SELECT  1 UNION ALL SELECT  2 UNION ALL SELECT  3 UNION ALL SELECT  4 UNION ALL SELECT  5 UNION ALL
    SELECT  6 UNION ALL SELECT  7 UNION ALL SELECT  8 UNION ALL SELECT  9 UNION ALL SELECT 10 UNION ALL
    SELECT 11 UNION ALL SELECT 12 UNION ALL SELECT 13 UNION ALL SELECT 14 UNION ALL SELECT 15 UNION ALL
    SELECT 16 UNION ALL SELECT 17 UNION ALL SELECT 18 UNION ALL SELECT 19 UNION ALL SELECT 20 UNION ALL
    SELECT 21 UNION ALL SELECT 22 UNION ALL SELECT 23 UNION ALL SELECT 24
)
INSERT INTO finance.Installments (
    [Id], [TransactionId], [InvoiceId],
    [Number], [Total], [Amount], [Currency],
    [Status], [LegacyOrigin], [CreatedAt], [UpdatedAt]
)
SELECT
    NEWID(),
    im.[NewTxId],
    im.[FinanceInvId],
    nb.[n],
    im.[InstallmentCount],
    CAST(ii.[InstallmentValue] AS DECIMAL(19,4)),
    'BRL',
    'scheduled',
    'InvoiceItem:' + CAST(ii.[Id] AS VARCHAR(36)) + ':' + CAST(nb.[n] AS VARCHAR(2)),
    GETUTCDATE(),
    GETUTCDATE()
FROM #InvoiceItemMap im
INNER JOIN dbo.InvoiceItem ii ON ii.[Id] = im.[LegacyItemId]
INNER JOIN Numbers nb ON nb.[n] <= im.[InstallmentCount]
WHERE im.[InstallmentCount] > 1
  AND NOT EXISTS (
        SELECT 1 FROM finance.Installments
        WHERE [LegacyOrigin] = 'InvoiceItem:' + CAST(ii.[Id] AS VARCHAR(36))
                             + ':' + CAST(nb.[n] AS VARCHAR(2))
);

DROP TABLE #InvoiceItemMap;
DROP TABLE #InvoiceMap;

-- ============================================================
-- RF-38a/b/c: Gate SQL — row counts per source table
-- ============================================================
DECLARE @srcCnt BIGINT, @dstCnt BIGINT;

-- RF-38a: TransactionItem count
SELECT @srcCnt = COUNT(*)
FROM dbo.TransactionItem ti
INNER JOIN dbo.[Transaction] t ON t.[Id] = ti.[TransactionId]
WHERE ti.[Active] = 1 AND t.[Active] = 1;

SELECT @dstCnt = COUNT(*) FROM finance.Transactions
WHERE [LegacyOrigin] LIKE 'TransactionItem:%';

IF @srcCnt <> @dstCnt
    RAISERROR('RF-38a: TransactionItem count mismatch — src=%d dst=%d', 16, 1, @srcCnt, @dstCnt);

-- RF-38b: InvoiceItem count
SELECT @srcCnt = COUNT(*)
FROM dbo.InvoiceItem ii
INNER JOIN dbo.Invoice inv ON inv.[Id] = ii.[InvoiceId]
WHERE ii.[Active] = 1 AND inv.[Active] = 1;

SELECT @dstCnt = COUNT(*) FROM finance.Transactions
WHERE [LegacyOrigin] LIKE 'InvoiceItem:%';

IF @srcCnt <> @dstCnt
    RAISERROR('RF-38b: InvoiceItem count mismatch — src=%d dst=%d', 16, 1, @srcCnt, @dstCnt);

-- RF-38c: Invoice count
SELECT @srcCnt = COUNT(*) FROM dbo.Invoice WHERE [Active] = 1;

SELECT @dstCnt = COUNT(*) FROM finance.Invoices
WHERE [LegacyOrigin] LIKE 'Invoice:%';

IF @srcCnt <> @dstCnt
    RAISERROR('RF-38c: Invoice count mismatch — src=%d dst=%d', 16, 1, @srcCnt, @dstCnt);

-- ============================================================
-- RF-38d: Gate SQL — sums per user / type (decimal exact)
-- ============================================================
DECLARE @mismatch INT;

-- TransactionItem income/expense sums per user
SELECT @mismatch = COUNT(*)
FROM (
    SELECT
        t.[UserId],
        CASE UPPER(ti.[Type]) WHEN 'INCOME' THEN 'income' ELSE 'expense' END AS [TxType],
        CAST(SUM(ti.[Value]) AS DECIMAL(19,4))                               AS [SrcSum]
    FROM dbo.TransactionItem ti
    INNER JOIN dbo.[Transaction] t ON t.[Id] = ti.[TransactionId]
    WHERE ti.[Active] = 1 AND t.[Active] = 1
    GROUP BY t.[UserId], CASE UPPER(ti.[Type]) WHEN 'INCOME' THEN 'income' ELSE 'expense' END
) src
FULL OUTER JOIN (
    SELECT
        [UserId],
        [TransactionType],
        CAST(SUM([Amount]) AS DECIMAL(19,4)) AS [DstSum]
    FROM finance.Transactions
    WHERE [LegacyOrigin] LIKE 'TransactionItem:%'
    GROUP BY [UserId], [TransactionType]
) dst ON dst.[UserId] = src.[UserId] AND dst.[TransactionType] = src.[TxType]
WHERE ISNULL(src.[SrcSum], 0) <> ISNULL(dst.[DstSum], 0);

IF @mismatch > 0
    RAISERROR('RF-38d: TransactionItem sum mismatch by user/type (%d rows differ)', 16, 1, @mismatch);

-- InvoiceItem TotalAmount sums per user
SELECT @mismatch = COUNT(*)
FROM (
    SELECT
        c.[UserId],
        CAST(SUM(CAST(ii.[TotalAmount] AS DECIMAL(19,4))) AS DECIMAL(19,4)) AS [SrcSum]
    FROM dbo.InvoiceItem ii
    INNER JOIN dbo.Invoice inv ON inv.[Id] = ii.[InvoiceId]
    INNER JOIN dbo.Card    c   ON c.[Id]   = inv.[CardId]
    WHERE ii.[Active] = 1 AND inv.[Active] = 1
    GROUP BY c.[UserId]
) src
FULL OUTER JOIN (
    SELECT
        [UserId],
        CAST(SUM([Amount]) AS DECIMAL(19,4)) AS [DstSum]
    FROM finance.Transactions
    WHERE [LegacyOrigin] LIKE 'InvoiceItem:%'
    GROUP BY [UserId]
) dst ON dst.[UserId] = src.[UserId]
WHERE ISNULL(src.[SrcSum], 0) <> ISNULL(dst.[DstSum], 0);

IF @mismatch > 0
    RAISERROR('RF-38d: InvoiceItem sum mismatch by user (%d rows differ)', 16, 1, @mismatch);

-- Installment N:N check — each installment_purchase must have exactly N installment rows
SELECT @mismatch = COUNT(*)
FROM (
    SELECT
        ii.[Id]                         AS ItemId,
        ISNULL(ii.[Installment], 1)     AS [Expected]
    FROM dbo.InvoiceItem ii
    INNER JOIN dbo.Invoice inv ON inv.[Id] = ii.[InvoiceId]
    WHERE ii.[Active] = 1 AND inv.[Active] = 1
      AND ISNULL(ii.[Installment], 1) > 1
) src
LEFT JOIN (
    SELECT
        ft.[LegacyOrigin],
        COUNT(*) AS [Actual]
    FROM finance.Installments inst
    INNER JOIN finance.Transactions ft ON ft.[Id] = inst.[TransactionId]
    WHERE ft.[LegacyOrigin] LIKE 'InvoiceItem:%'
      AND ft.[LegacyOrigin] NOT LIKE 'InvoiceItem:%:%'
    GROUP BY ft.[LegacyOrigin]
) dst ON dst.[LegacyOrigin] = 'InvoiceItem:' + CAST(src.[ItemId] AS VARCHAR(36))
WHERE src.[Expected] <> ISNULL(dst.[Actual], 0);

IF @mismatch > 0
    RAISERROR('RF-38d: Installment N:N mismatch (%d invoice items)', 16, 1, @mismatch);

-- ============================================================
-- Audit: mark DML complete (Status = dml_ok)
-- ============================================================
UPDATE finance.MigrationAudit
SET
    [Status]      = 'dml_ok',
    [FinishedAt]  = GETUTCDATE(),
    [RowsRead]    =
        (SELECT COUNT(*) FROM dbo.TransactionItem WHERE Active = 1) +
        (SELECT COUNT(*) FROM dbo.InvoiceItem     WHERE Active = 1) +
        (SELECT COUNT(*) FROM dbo.Invoice         WHERE Active = 1),
    [RowsWritten] =
        (SELECT COUNT(*) FROM finance.Transactions  WHERE LegacyOrigin IS NOT NULL) +
        (SELECT COUNT(*) FROM finance.Invoices      WHERE LegacyOrigin IS NOT NULL) +
        (SELECT COUNT(*) FROM finance.Installments  WHERE LegacyOrigin IS NOT NULL)
WHERE [Id] = @auditId;

-- ============================================================
-- Table: finance.MigrationAudit  (Assumption 4 / RF-38)
-- ============================================================
CREATE TABLE finance.MigrationAudit (
    [Id]           UNIQUEIDENTIFIER NOT NULL,
    [StartedAt]    DATETIME2        NOT NULL,
    [FinishedAt]   DATETIME2        NULL,
    [SourceTable]  VARCHAR(64)      NOT NULL,
    [RowsRead]     BIGINT           NOT NULL DEFAULT 0,
    [RowsWritten]  BIGINT           NOT NULL DEFAULT 0,
    [Status]       VARCHAR(16)      NOT NULL,
    CONSTRAINT PK_MigrationAudit PRIMARY KEY ([Id])
);
