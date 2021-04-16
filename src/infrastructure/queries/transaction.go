package queries

const (
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
	GetTransactionByDate = `SELECT 
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
							WHERE
								[Active] = 1
								AND [Date] BETWEEN CONVERT(DATETIME, @startDate)
								AND CONVERT(DATETIME, @endDate)`
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
							AND [Active] = 1
							ORDER BY [Type], [Value]`
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
							AND [TransactionId] = @transactionId
							AND [Active] = 1`
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
								[Id] = @id
							AND [TransactionId] = @transactionId`
)
