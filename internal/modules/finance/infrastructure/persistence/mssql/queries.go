package mssql

// Transaction queries
const (
	addTransaction = `INSERT INTO dbo.FinanceTransactions
		([Id],[UserId],[Description],[Amount],[Currency],[OccurredAt],[TransactionType],
		 [PaymentMethod],[CardId],[CategoryId],[SubcategoryId],[OriginalTransactionId],
		 [LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt])
		VALUES
		(@id,@userId,@description,@amount,@currency,@occurredAt,@transactionType,
		 @paymentMethod,@cardId,@categoryId,@subcategoryId,@originalTransactionId,
		 @legacyOrigin,@createdAt,@updatedAt,@deletedAt)`

	updateTransaction = `UPDATE dbo.FinanceTransactions SET
		[Description]          = @description,
		[Amount]               = @amount,
		[OccurredAt]           = @occurredAt,
		[TransactionType]      = @transactionType,
		[PaymentMethod]        = @paymentMethod,
		[CardId]               = @cardId,
		[CategoryId]           = @categoryId,
		[SubcategoryId]        = @subcategoryId,
		[OriginalTransactionId]= @originalTransactionId,
		[UpdatedAt]            = @updatedAt,
		[DeletedAt]            = @deletedAt
	WHERE [Id] = @id AND [UserId] = @userId AND [DeletedAt] IS NULL`

	getTransactionByID = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([UserId] AS CHAR(36)),
		[Description],[Amount],[Currency],[OccurredAt],[TransactionType],[PaymentMethod],
		CAST([CardId] AS CHAR(36)),
		CAST([CategoryId] AS CHAR(36)),
		CAST([SubcategoryId] AS CHAR(36)),
		CAST([OriginalTransactionId] AS CHAR(36)),
		[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceTransactions (NOLOCK)
	WHERE [UserId] = @userId AND [Id] = @id AND [DeletedAt] IS NULL`

	softDeleteTransaction = `UPDATE dbo.FinanceTransactions SET
		[DeletedAt] = @deletedAt, [UpdatedAt] = @updatedAt
	WHERE [Id] = @id AND [UserId] = @userId AND [DeletedAt] IS NULL`

	hasActiveRefundFor = `SELECT COUNT(1) FROM dbo.FinanceTransactions (NOLOCK)
	WHERE [OriginalTransactionId] = @originalId
	  AND [TransactionType] = 'refund'
	  AND [DeletedAt] IS NULL
	  AND [UserId] = @userId`

	// base SELECT columns reused in list + count
	transactionListSelect = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([UserId] AS CHAR(36)),
		[Description],[Amount],[Currency],[OccurredAt],[TransactionType],[PaymentMethod],
		CAST([CardId] AS CHAR(36)),
		CAST([CategoryId] AS CHAR(36)),
		CAST([SubcategoryId] AS CHAR(36)),
		CAST([OriginalTransactionId] AS CHAR(36)),
		[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceTransactions (NOLOCK)`

	// ORDER BY for list
	transactionListOrder = ` ORDER BY [OccurredAt] DESC, [CreatedAt] DESC`

	// OFFSET/FETCH appended by List
	transactionListPaging = ` OFFSET @offset ROWS FETCH NEXT @size ROWS ONLY`

	sumForSummary = `SELECT
		ISNULL(SUM(CASE WHEN t.[TransactionType] = 'income'  THEN t.[Amount] ELSE 0 END), 0),
		ISNULL(SUM(CASE WHEN t.[TransactionType] = 'expense' THEN t.[Amount] ELSE 0 END), 0),
		ISNULL(SUM(CASE WHEN t.[TransactionType] = 'refund' AND o.[TransactionType] = 'income' THEN t.[Amount] ELSE 0 END), 0),
		ISNULL(SUM(CASE WHEN t.[TransactionType] = 'refund' AND (o.[TransactionType] IS NULL OR o.[TransactionType] != 'income') THEN t.[Amount] ELSE 0 END), 0)
	FROM dbo.FinanceTransactions t (NOLOCK)
	LEFT JOIN dbo.FinanceTransactions o (NOLOCK) ON o.[Id] = t.[OriginalTransactionId]
	WHERE t.[UserId] = @userId AND t.[DeletedAt] IS NULL
	  AND t.[OccurredAt] >= @from AND t.[OccurredAt] < @to`
)

// Invoice queries
const (
	addInvoice = `INSERT INTO dbo.FinanceInvoices
		([Id],[UserId],[CardId],[State],[CycleStart],[CycleEnd],[ClosingDate],[DueDate],
		 [Total],[Currency],[PaidAt],[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt])
		VALUES
		(@id,@userId,@cardId,@state,@cycleStart,@cycleEnd,@closingDate,@dueDate,
		 @total,@currency,@paidAt,@legacyOrigin,@createdAt,@updatedAt,@deletedAt)`

	updateInvoice = `UPDATE dbo.FinanceInvoices SET
		[State]      = @state,
		[Total]      = @total,
		[PaidAt]     = @paidAt,
		[UpdatedAt]  = @updatedAt,
		[DeletedAt]  = @deletedAt
	WHERE [Id] = @id AND [UserId] = @userId AND [DeletedAt] IS NULL`

	getInvoiceByID = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([UserId] AS CHAR(36)),
		CAST([CardId] AS CHAR(36)),
		[State],[CycleStart],[CycleEnd],[ClosingDate],[DueDate],[Total],[Currency],
		[PaidAt],[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceInvoices (NOLOCK)
	WHERE [UserId] = @userId AND [Id] = @id AND [DeletedAt] IS NULL`

	// find open invoice for card with a specific closing date (UPDLOCK prevents race)
	getOpenInvoiceForCardClosing = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([UserId] AS CHAR(36)),
		CAST([CardId] AS CHAR(36)),
		[State],[CycleStart],[CycleEnd],[ClosingDate],[DueDate],[Total],[Currency],
		[PaidAt],[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceInvoices WITH (UPDLOCK, HOLDLOCK)
	WHERE [UserId] = @userId AND [CardId] = @cardId AND [State] = 'open'
	  AND CAST([ClosingDate] AS DATE) = CAST(@closingDate AS DATE)
	  AND [DeletedAt] IS NULL`

	nextOpenInvoiceForCard = `SELECT TOP 1
		CAST([Id] AS CHAR(36)),
		CAST([UserId] AS CHAR(36)),
		CAST([CardId] AS CHAR(36)),
		[State],[CycleStart],[CycleEnd],[ClosingDate],[DueDate],[Total],[Currency],
		[PaidAt],[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceInvoices (NOLOCK)
	WHERE [UserId] = @userId AND [CardId] = @cardId AND [State] = 'open'
	  AND [DeletedAt] IS NULL
	ORDER BY [CycleEnd] ASC`

	sumPaidInPeriod = `SELECT ISNULL(SUM([Total]), 0)
	FROM dbo.FinanceInvoices (NOLOCK)
	WHERE [UserId] = @userId AND [State] = 'paid' AND [DeletedAt] IS NULL
	  AND [PaidAt] >= @from AND [PaidAt] < @to`

	sumOpenForUser = `SELECT ISNULL(SUM([Total]), 0)
	FROM dbo.FinanceInvoices (NOLOCK)
	WHERE [UserId] = @userId AND [State] = 'open' AND [DeletedAt] IS NULL
	  AND [CycleEnd] >= @from AND [CycleEnd] < @to`

	invoiceListSelect = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([UserId] AS CHAR(36)),
		CAST([CardId] AS CHAR(36)),
		[State],[CycleStart],[CycleEnd],[ClosingDate],[DueDate],[Total],[Currency],
		[PaidAt],[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceInvoices (NOLOCK)`

	invoiceListOrder  = ` ORDER BY [CycleEnd] DESC`
	invoiceListPaging = ` OFFSET @offset ROWS FETCH NEXT @size ROWS ONLY`
)

// Installment queries
const (
	addInstallment = `INSERT INTO dbo.FinanceInstallments
		([Id],[TransactionId],[InvoiceId],[Number],[Total],[Amount],[Currency],
		 [Status],[LegacyOrigin],[CreatedAt],[UpdatedAt],[DeletedAt])
		VALUES
		(@id,@transactionId,@invoiceId,@number,@total,@amount,@currency,
		 @status,@legacyOrigin,@createdAt,@updatedAt,@deletedAt)`

	updateInstallment = `UPDATE dbo.FinanceInstallments SET
		[InvoiceId]   = @invoiceId,
		[Status]      = @status,
		[UpdatedAt]   = @updatedAt,
		[DeletedAt]   = @deletedAt
	WHERE [Id] = @id AND [DeletedAt] IS NULL`

	// softDeleteInstallmentsByTransaction soft-deletes every active installment of
	// the given transaction. The JOIN with dbo.FinanceTransactions enforces user
	// authorization (t.[UserId] = @userId). It intentionally does NOT filter
	// t.[DeletedAt] IS NULL: DeleteTransaction.Execute soft-deletes the parent
	// row first inside the same database.Do transaction, so a DeletedAt-IS-NULL
	// predicate here would exclude the just-soft-deleted parent and the UPDATE
	// would affect zero installments (RF-44/RF-55 regression).
	softDeleteInstallmentsByTransaction = `UPDATE i SET
		i.[DeletedAt] = @deletedAt, i.[UpdatedAt] = @updatedAt, i.[Status] = 'refunded'
	FROM dbo.FinanceInstallments i
	INNER JOIN dbo.FinanceTransactions t ON t.[Id] = i.[TransactionId]
	WHERE i.[TransactionId] = @transactionId
	  AND i.[DeletedAt] IS NULL
	  AND t.[UserId] = @userId`

	listInstallmentsByInvoice = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([TransactionId] AS CHAR(36)),
		CAST([InvoiceId] AS CHAR(36)),
		[Number],[Total],[Amount],[Currency],[Status],[LegacyOrigin],
		[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceInstallments (NOLOCK)
	WHERE [InvoiceId] = @invoiceId AND [DeletedAt] IS NULL
	ORDER BY [Number] ASC`

	listInstallmentsByTransaction = `SELECT
		CAST([Id] AS CHAR(36)),
		CAST([TransactionId] AS CHAR(36)),
		CAST([InvoiceId] AS CHAR(36)),
		[Number],[Total],[Amount],[Currency],[Status],[LegacyOrigin],
		[CreatedAt],[UpdatedAt],[DeletedAt]
	FROM dbo.FinanceInstallments (NOLOCK)
	WHERE [TransactionId] = @transactionId AND [DeletedAt] IS NULL
	ORDER BY [Number] ASC`

	// hasClosedOrPaidForTransaction returns >0 when any non-deleted installment of the
	// transaction is bound to an invoice whose state is 'closed' or 'paid'.
	// Reads the parent invoice state directly (RF-46 keeps installment.status=='scheduled'
	// while the invoice is closed-but-unpaid, so checking installment.status alone
	// would miss the closed case required by RF-11/RF-14/RF-44).
	// Invoices is read WITHOUT (NOLOCK) on purpose: inv.[State] is load-bearing for
	// authorization (RF-44/RF-55) and a dirty read of an in-flight close that later
	// rolls back would block legitimate updates/deletes.
	hasClosedOrPaidForTransaction = `SELECT COUNT(1)
	FROM dbo.FinanceInstallments i (NOLOCK)
	INNER JOIN dbo.FinanceInvoices inv ON inv.[Id] = i.[InvoiceId]
	WHERE i.[TransactionId] = @transactionId
	  AND i.[DeletedAt] IS NULL
	  AND inv.[DeletedAt] IS NULL
	  AND inv.[State] IN ('closed','paid')`

	sumByMonthCompetence = `SELECT ISNULL(SUM(i.[Amount]), 0)
	FROM dbo.FinanceInstallments i (NOLOCK)
	INNER JOIN dbo.FinanceInvoices inv (NOLOCK) ON inv.[Id] = i.[InvoiceId]
	WHERE inv.[UserId] = @userId
	  AND inv.[CycleStart] >= @from AND inv.[CycleStart] < @to
	  AND i.[DeletedAt] IS NULL AND inv.[DeletedAt] IS NULL`
)
