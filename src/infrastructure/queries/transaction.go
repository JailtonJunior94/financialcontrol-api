package queries

const (
	GetByEmail = `SELECT
					CAST([Id] AS CHAR(36)) [Id],
					[Name],
					[Email],
					[Password],
					[CreatedAt],
					[UpdatedAt],
					[Active]
				FROM
					dbo.[User] (NOLOCK)
				WHERE [Email] = @email
				AND [Active] = 1`
	GetTransactions = `SELECT
						CAST([Id] AS CHAR(36)) [Id],
						CAST([UserId] AS CHAR(36)) [UserId],
						[Date],
						[Total],
						[Income],
						[Outcome],
						[CreatedAt],
						[UpdatedAt],
						[Active]
					FROM
						dbo.[Transaction] (NOLOCK)
					WHERE [UserId] = @userId
					AND [Active] = 1
					ORDER BY [Date]`
	GetTransactionById = `SELECT 
							CAST([Id] AS CHAR(36)) [Id],
							CAST([UserId] AS CHAR(36)) [UserId],
							[Date],
							[Total],
							[Income],
							[Outcome],
							[CreatedAt],
							[UpdatedAt],
							[Active]
						FROM 
							dbo.[Transaction] (NOLOCK)
						WHERE [Id] = @id 
						AND [UserId] = @userId`
	GetItemByTransactionId = `SELECT 
								CAST([Id] AS CHAR(36)) [Id],
								CAST([TransactionId] AS CHAR(36)) [TransactionId],
								[Title],
								[Value],
								[Type],
								[CreatedAt],
								[UpdatedAt],
								[Active]
							FROM 
								dbo.[TransactionItem] (NOLOCK)
							WHERE [TransactionId] = @transactionId
							AND [Active] = 1`
	GetTransactionItemsById = `SELECT
								CAST([Id] AS CHAR(36)) [Id],
								CAST([TransactionId] AS CHAR(36)) [TransactionId],
								[Title],
								[Value],
								[Type],
								[CreatedAt],
								[UpdatedAt],
								[Active]
							FROM
								dbo.[TransactionItem] (NOLOCK)
							WHERE [Id] = @id
							AND [Active] = 1`
	AddUser           = `INSERT INTO dbo.[User] VALUES (@id, @name, @email, @password, @createdAt, @updatedAt, @active)`
	AddTransaction    = `INSERT INTO dbo.[Transaction] VALUES (@id, @userId, @date, @total, @income, @outcome, @createdAt, @updatedAt, @active)`
	UpdateTransaction = `UPDATE
							dbo.[Transaction]
						SET
							[Total] = @total,
							[Income] = @income,
							[Outcome] = @outcome,
							[UpdatedAt] = @updatedAt
						WHERE
							[Id] = @id`
	AddTransactionItem    = `INSERT INTO dbo.[TransactionItem] VALUES (@id, @transactionId, @title, @value, @type, @createdAt, @updatedAt, @active)`
	UpdateTransactionItem = `UPDATE
								dbo.[TransactionItem]
							SET
								[Title] = @title,
								[Value] = @value,
								[Type] = @type,
								[UpdatedAt] = @updatedAt,
								[Active] = @active
							WHERE
								[Id] = @id`
)