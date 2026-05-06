-- Migration: create Category table with partial unique indexes for active rows.
-- Soft delete is represented by DeletedAt; uniqueness applies only to active rows.
-- Subcategory depth (max 2) is enforced at the application layer (ADR-001).

IF NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'Category'
)
CREATE TABLE dbo.[Category] (
    [Id]        CHAR(36)      NOT NULL PRIMARY KEY,
    [UserId]    CHAR(36)      NOT NULL,
    [ParentId]  CHAR(36)      NULL,
    [Name]      NVARCHAR(100) NOT NULL,
    [Color]     VARCHAR(32)   NOT NULL,
    [Icon]      VARCHAR(64)   NOT NULL,
    [CreatedAt] DATETIME2     NOT NULL,
    [UpdatedAt] DATETIME2     NOT NULL,
    [DeletedAt] DATETIME2     NULL,
    CONSTRAINT FK_Category_Parent FOREIGN KEY ([ParentId]) REFERENCES dbo.[Category]([Id])
);

IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'IX_Category_User_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE INDEX IX_Category_User_Active
    ON dbo.[Category]([UserId], [DeletedAt])
    INCLUDE([ParentId], [Name]);

IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'IX_Category_User_Parent_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE INDEX IX_Category_User_Parent_Active
    ON dbo.[Category]([UserId], [ParentId], [DeletedAt]);

IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'UX_Category_Root_Name_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE UNIQUE INDEX UX_Category_Root_Name_Active
    ON dbo.[Category]([UserId], [Name])
    WHERE [DeletedAt] IS NULL AND [ParentId] IS NULL;

IF NOT EXISTS (
    SELECT 1 FROM sys.indexes
    WHERE name = 'UX_Category_Sub_Name_Active' AND object_id = OBJECT_ID('dbo.Category')
)
CREATE UNIQUE INDEX UX_Category_Sub_Name_Active
    ON dbo.[Category]([UserId], [ParentId], [Name])
    WHERE [DeletedAt] IS NULL AND [ParentId] IS NOT NULL;
