-- grant_obs_reader.sql — creates a minimal read-only SQL login and user
-- for the otelcol-contrib sqlserverreceiver.
--
-- Permissions granted:
--   VIEW SERVER STATE  — required for sys.dm_os_wait_stats, sys.dm_exec_*
--   VIEW DATABASE STATE — required for sys.query_store_* and database DMVs
--
-- Run as sysadmin against the target database (e.g. financial_control).
-- Script is idempotent: each CREATE is guarded by IF NOT EXISTS.

-- 1. Create SQL login (server-level) if it does not exist.
IF NOT EXISTS (
    SELECT 1 FROM sys.server_principals WHERE name = N'obs_reader'
)
BEGIN
    -- Use a strong password managed via your secret manager in production.
    CREATE LOGIN [obs_reader] WITH PASSWORD = N'$(OBS_READER_PASSWORD)';
    PRINT 'Login obs_reader created.';
END
ELSE
BEGIN
    PRINT 'Login obs_reader already exists — skipping CREATE LOGIN.';
END

-- 2. Grant VIEW SERVER STATE at the server level (needed for DMVs).
IF NOT EXISTS (
    SELECT 1
    FROM   sys.server_permissions sp
    JOIN   sys.server_principals  pr ON sp.grantee_principal_id = pr.principal_id
    WHERE  pr.name        = N'obs_reader'
    AND    sp.permission_name = 'VIEW SERVER STATE'
    AND    sp.state_desc  = 'GRANT'
)
BEGIN
    GRANT VIEW SERVER STATE TO [obs_reader];
    PRINT 'VIEW SERVER STATE granted to obs_reader.';
END
ELSE
BEGIN
    PRINT 'VIEW SERVER STATE already granted — skipping.';
END

-- 3. Create database user mapped to the login (run in context of target database).
IF NOT EXISTS (
    SELECT 1 FROM sys.database_principals WHERE name = N'obs_reader'
)
BEGIN
    CREATE USER [obs_reader] FOR LOGIN [obs_reader];
    PRINT 'Database user obs_reader created.';
END
ELSE
BEGIN
    PRINT 'Database user obs_reader already exists — skipping CREATE USER.';
END

-- 4. Grant VIEW DATABASE STATE (needed for sys.query_store_* and db-level DMVs).
IF NOT EXISTS (
    SELECT 1
    FROM   sys.database_permissions dp
    JOIN   sys.database_principals  pr ON dp.grantee_principal_id = pr.principal_id
    WHERE  pr.name            = N'obs_reader'
    AND    dp.permission_name = 'VIEW DATABASE STATE'
    AND    dp.state_desc      = 'GRANT'
)
BEGIN
    GRANT VIEW DATABASE STATE TO [obs_reader];
    PRINT 'VIEW DATABASE STATE granted to obs_reader.';
END
ELSE
BEGIN
    PRINT 'VIEW DATABASE STATE already granted — skipping.';
END

PRINT 'grant_obs_reader.sql completed successfully.';
