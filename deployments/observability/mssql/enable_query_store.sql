-- enable_query_store.sql — idempotent.
-- Requires SQL Server 2017+ (ProductMajorVersion >= 14).
-- Run against the target database (e.g. financial_control) with sufficient privileges.
-- Second execution is a no-op: the IF EXISTS guard prevents duplicate ALTER DATABASE.

IF CAST(SERVERPROPERTY('ProductMajorVersion') AS INT) < 14
BEGIN
    RAISERROR('Query Store requires SQL Server 2017+ (ProductMajorVersion >= 14). Current version: %s',
              16, 1, CAST(SERVERPROPERTY('ProductVersion') AS NVARCHAR(20)));
    RETURN;
END

DECLARE @db SYSNAME = DB_NAME();

-- Only ALTER DATABASE when Query Store is not already in READ_WRITE state.
IF EXISTS (
    SELECT 1
    FROM   sys.database_query_store_options
    WHERE  actual_state_desc <> 'READ_WRITE'
)
BEGIN
    EXEC (
        'ALTER DATABASE [' + @db + '] SET QUERY_STORE = ON (
            OPERATION_MODE         = READ_WRITE,
            CLEANUP_POLICY         = (STALE_QUERY_THRESHOLD_DAYS = 30),
            DATA_FLUSH_INTERVAL_SECONDS = 900,
            INTERVAL_LENGTH_MINUTES = 15,
            MAX_STORAGE_SIZE_MB    = 1024,
            QUERY_CAPTURE_MODE     = AUTO,
            SIZE_BASED_CLEANUP_MODE = AUTO
        );'
    );
    PRINT 'Query Store enabled on database: ' + @db;
END
ELSE
BEGIN
    PRINT 'Query Store already in READ_WRITE state on database: ' + @db + ' — no changes made.';
END
