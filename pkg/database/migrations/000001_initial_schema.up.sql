-- Migration: create the initial legacy Financial Control schema on MSSQL.
-- The target database is selected by MSSQL_CONNECTION_STRING; object names stay
-- scoped to dbo without a physical database prefix.
-- Idempotency is delegated to dbo.schema_migrations; defensive guards removed.

CREATE TABLE dbo.Bill (
    Id             uniqueidentifier NOT NULL,
    [Date]         datetime2        NOT NULL,
    Total          decimal(18,2)    NULL,
    SixtyPercent   decimal(18,2)    NULL,
    FortyPercent   decimal(18,2)    NULL,
    CreatedAt      datetime2        NOT NULL,
    UpdatedAt      datetime2        NULL,
    Active         bit              NOT NULL,
    CONSTRAINT PK__Bill__3214EC07FCC438A1 PRIMARY KEY (Id)
);

CREATE TABLE dbo.Category (
    Id         uniqueidentifier                               NOT NULL,
    Name       varchar(100) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    [Sequence] int                                            NOT NULL,
    CreatedAt  datetime2                                      NOT NULL,
    UpdatedAt  datetime2                                      NULL,
    Active     bit                                            NOT NULL,
    CONSTRAINT PK_Category PRIMARY KEY (Id)
);

CREATE TABLE dbo.Flag (
    Id        uniqueidentifier                               NOT NULL,
    Name      varchar(100) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    CreatedAt datetime2                                      NOT NULL,
    UpdatedAt datetime2                                      NULL,
    Active    bit                                            NOT NULL,
    CONSTRAINT PK_Flag PRIMARY KEY (Id)
);

CREATE TABLE dbo.[User] (
    Id        uniqueidentifier                               NOT NULL,
    Name      varchar(50) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    Email     varchar(100) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    Password  varchar(300) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    CreatedAt datetime2                                      NOT NULL,
    UpdatedAt datetime2                                      NULL,
    Active    bit                                            NOT NULL,
    CONSTRAINT PK_User PRIMARY KEY (Id)
);

CREATE TABLE dbo.BillItem (
    Id        uniqueidentifier                                NOT NULL,
    BillId    uniqueidentifier                                NOT NULL,
    Title     varchar(100) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
    Value     decimal(18,2)                                   NOT NULL,
    CreatedAt datetime2                                       NOT NULL,
    UpdatedAt datetime2                                       NULL,
    Active    bit                                             NOT NULL,
    CONSTRAINT PK__BillItem__3214EC07D759346A PRIMARY KEY (Id),
    CONSTRAINT FK__BillItem__BillId__71D1E811 FOREIGN KEY (BillId) REFERENCES dbo.Bill(Id)
);

CREATE TABLE dbo.Card (
    Id             uniqueidentifier                                NOT NULL,
    UserId         uniqueidentifier                                NOT NULL,
    FlagId         uniqueidentifier                                NOT NULL,
    Name           varchar(50) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    Number         varchar(50) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    Description    varchar(100) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    ClosingDay     int                                             NOT NULL,
    ExpirationDate datetime2                                       NOT NULL,
    CreatedAt      datetime2                                       NOT NULL,
    UpdatedAt      datetime2                                       NULL,
    Active         bit                                             NOT NULL,
    CONSTRAINT PK_Card PRIMARY KEY (Id),
    CONSTRAINT FK_Card_Flag_FlagId FOREIGN KEY (FlagId) REFERENCES dbo.Flag(Id) ON DELETE CASCADE,
    CONSTRAINT FK_Card_User_UserId FOREIGN KEY (UserId) REFERENCES dbo.[User](Id) ON DELETE CASCADE
);

CREATE TABLE dbo.Invoice (
    Id        uniqueidentifier NOT NULL,
    CardId    uniqueidentifier NOT NULL,
    [Date]    datetime2        NOT NULL,
    Total     float            NOT NULL,
    CreatedAt datetime2        NOT NULL,
    UpdatedAt datetime2        NULL,
    Active    bit              NOT NULL,
    CONSTRAINT PK_Invoice PRIMARY KEY (Id),
    CONSTRAINT FK_Invoice_Card_CardId FOREIGN KEY (CardId) REFERENCES dbo.Card(Id) ON DELETE CASCADE
);

CREATE TABLE dbo.InvoiceItem (
    Id               uniqueidentifier                                 NOT NULL,
    InvoiceId        uniqueidentifier                                 NOT NULL,
    CategoryId       uniqueidentifier                                 NOT NULL,
    PurchaseDate     datetime2                                        NOT NULL,
    Description      varchar(800) COLLATE SQL_Latin1_General_CP1_CI_AS NOT NULL,
    TotalAmount      float                                            NOT NULL,
    Installment      int                                              NULL,
    InstallmentValue float                                            NOT NULL,
    Tags             varchar(800) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
    CreatedAt        datetime2                                        NOT NULL,
    UpdatedAt        datetime2                                        NULL,
    Active           bit                                              NOT NULL,
    InvoiceControl   bigint                                           NULL,
    CONSTRAINT PK_InvoiceItem PRIMARY KEY (Id),
    CONSTRAINT FK_InvoiceItem_Category_CategoryId FOREIGN KEY (CategoryId) REFERENCES dbo.Category(Id) ON DELETE CASCADE,
    CONSTRAINT FK_InvoiceItem_Invoice_InvoiceId FOREIGN KEY (InvoiceId) REFERENCES dbo.Invoice(Id) ON DELETE CASCADE
);

CREATE TABLE dbo.[Transaction] (
    Id        uniqueidentifier NOT NULL,
    UserId    uniqueidentifier NOT NULL,
    [Date]    datetime2        NOT NULL,
    Total     decimal(18,2)    NULL,
    Income    decimal(18,2)    NULL,
    Outcome   decimal(18,2)    NULL,
    CreatedAt datetime2        NOT NULL,
    UpdatedAt datetime2        NULL,
    Active    bit              NOT NULL,
    CONSTRAINT PK_Transaction PRIMARY KEY (Id),
    CONSTRAINT FK_Transaction_User_UserId FOREIGN KEY (UserId) REFERENCES dbo.[User](Id) ON DELETE CASCADE
);

CREATE TABLE dbo.TransactionItem (
    Id            uniqueidentifier                                 NOT NULL,
    TransactionId uniqueidentifier                                 NOT NULL,
    Title         varchar(100) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
    Value         decimal(18,2)                                    NOT NULL,
    [Type]        varchar(10) COLLATE SQL_Latin1_General_CP1_CI_AS NULL,
    CreatedAt     datetime2                                        NOT NULL,
    UpdatedAt     datetime2                                        NULL,
    Active        bit                                              NOT NULL,
    IsPaid        bit                                              NULL,
    CONSTRAINT PK_TransactionItem PRIMARY KEY (Id),
    CONSTRAINT FK_TransactionItem_Transaction_TransactionId FOREIGN KEY (TransactionId) REFERENCES dbo.[Transaction](Id) ON DELETE CASCADE
);

ALTER TABLE dbo.TransactionItem WITH NOCHECK
ADD CONSTRAINT CK_TYPE CHECK (([Type] = 'INCOME' OR [Type] = 'OUTCOME'));
