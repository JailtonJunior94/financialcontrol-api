package infrastructure

const (
	getTransactions = `SELECT
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
	getTransactionByDate = `SELECT 
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
								AND [UserId] = @userId
								AND [Date] BETWEEN CONVERT(DATETIME, @startDate)
								AND CONVERT(DATETIME, @endDate)`
	getTransactionById = `SELECT 
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
	getItemByTransactionId = `SELECT 
								CAST([Id] AS CHAR(36)) [Id],
								CAST([TransactionId] AS CHAR(36)) [TransactionId],
								[Title],
								[Value],
								[Type],
								[CreatedAt],
								[UpdatedAt],
								[IsPaid],
								[Active]
							FROM 
								dbo.[TransactionItem] (NOLOCK)
							WHERE [TransactionId] = @transactionId
							AND [Active] = 1
							ORDER BY [Type], [Value] DESC`
	getTransactionItemsById = `SELECT
								CAST([Id] AS CHAR(36)) [Id],
								CAST([TransactionId] AS CHAR(36)) [TransactionId],
								[Title],
								[Value],
								[Type],
								[CreatedAt],
								[UpdatedAt],
								[IsPaid],
								[Active]
							FROM
								dbo.[TransactionItem] (NOLOCK)
							WHERE [Id] = @id
							AND [TransactionId] = @transactionId
							AND [Active] = 1`
	addTransaction    = `INSERT INTO dbo.[Transaction] VALUES (@id, @userId, @date, @total, @income, @outcome, @createdAt, @updatedAt, @active)`
	updateTransaction = `UPDATE
							dbo.[Transaction]
						SET
							[Total] = @total,
							[Income] = @income,
							[Outcome] = @outcome,
							[UpdatedAt] = @updatedAt
						WHERE
							[Id] = @id`
	addTransactionItem    = `INSERT INTO dbo.[TransactionItem] VALUES (@id, @transactionId, @title, @value, @type, @createdAt, @updatedAt, @active, @isPaid)`
	updateTransactionItem = `UPDATE
								dbo.[TransactionItem]
							SET
								[Title] = @title,
								[Value] = @value,
								[Type] = @type,
								[UpdatedAt] = @updatedAt,
								[Active] = @active,
								[IsPaid] = @isPaid
							WHERE
								[Id] = @id
								AND [TransactionId] = @transactionId`
)
