-- Migration: drop legacy dbo tables after finance module smoke validation.
-- This migration is IRREVERSIBLE (decision C2.a / D4.a / RF-39 / ADR-006).
-- Recovery: restore from pre-migration backup snapshot.
--
-- Guard (RF-39 / defense in depth): the Go smoke hook (finance_smoke.go) must have
-- written Status='smoke_ok' to dbo.FinanceMigrationAudit before this file executes.
-- If the guard fires, the entire migration batch is aborted and the legacy tables
-- are preserved intact.

IF NOT EXISTS (
    SELECT 1 FROM dbo.FinanceMigrationAudit WHERE [Status] = 'smoke_ok'
)
BEGIN
    -- THROW aborts the batch unconditionally, so the legacy DROP statements
    -- below cannot run when the smoke gate has not been satisfied —
    -- independent of XACT_ABORT or the migration driver's error handling.
    THROW 50001,
        '000003: smoke validation not confirmed — Status=''smoke_ok'' required in dbo.FinanceMigrationAudit. Run cmd/migration with --smoke=finance flag.',
        1;
END;

-- ============================================================
-- Drop child tables first (removes inbound FK references)
-- ============================================================
IF OBJECT_ID('dbo.InvoiceItem', 'U') IS NOT NULL
    DROP TABLE dbo.InvoiceItem;

IF OBJECT_ID('dbo.TransactionItem', 'U') IS NOT NULL
    DROP TABLE dbo.TransactionItem;

-- ============================================================
-- Drop parent tables (outbound FK constraints dropped with the table)
-- ============================================================
IF OBJECT_ID('dbo.Invoice', 'U') IS NOT NULL
    DROP TABLE dbo.Invoice;

IF OBJECT_ID('dbo.Transaction', 'U') IS NOT NULL
    DROP TABLE dbo.[Transaction];
