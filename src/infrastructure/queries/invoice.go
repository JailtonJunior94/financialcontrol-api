package queries

const (
	GetInvoiceByCardId = `SELECT
							CAST([Id] AS CHAR(36)) [Id],
							CAST([CardId] AS CHAR(36)) [CardId],
							[Date],
							[Total],
							[CreatedAt],
							[UpdatedAt],
							[Active]
						FROM dbo.[Invoice] (NOLOCK)
						WHERE [CardId] = @cardId`
	GetInvoiceByDate = `SELECT
							CAST([Id] AS CHAR(36)) [Id],
							CAST([CardId] AS CHAR(36)) [CardId],
							[Date],
							[Total],
							[CreatedAt],
							[UpdatedAt],
							[Active]
						FROM dbo.[Invoice] (NOLOCK)
						WHERE [CardId] = @cardId	
						AND [Date] BETWEEN CONVERT(DATETIME, @startDate)
						AND CONVERT(DATETIME, @endDate)`
	GetInvoiceItemByInvoiceId = `SELECT
									CAST(I.[Id] AS CHAR(36)) [Id],
									CAST(I.[InvoiceId] AS CHAR(36)) [InvoiceId],
									CAST(I.[CategoryId] AS CHAR(36)) [CategoryId],
									I.[PurchaseDate],
									I.[Description],
									I.[TotalAmount],
									I.[Installment],
									I.[InstallmentValue],
									I.[Tags],
									I.[CreatedAt],
									I.[UpdatedAt],
									I.[Active],
									C.[Id],
									C.[Name],
									C.[Active]
								FROM
									dbo.[InvoiceItem] (NOLOCK) I
									INNER JOIN dbo.[Category] (NOLOCK) C ON C.Id = i.CategoryId
								WHERE
									InvoiceId = @invoiceId`
	AddInvoice     = `INSERT INTO dbo.[Invoice] VALUES (@id, @cardId, @date, @total, @createdAt, @updatedAt, @active)`
	UpdateInvoice  = `UPDATE dbo.[Invoice] SET Total = @total WHERE Id = @id`
	AddInvoiceItem = `INSERT INTO
							dbo.[InvoiceItem]
						VALUES
							(
							@id,
							@invoiceId,
							@categoryId,
							@purchaseDate,
							@description,
							@totalAmount,
							@installment,
							@installmentValue,
							@tags,
							@createdAt,
							@updatedAt,
							@active
							)`
)
