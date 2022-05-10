package queries

const (
	GetInvoiceByCardId = `SELECT
							CAST(I.[Id] AS CHAR(36)) [Id],
							CAST(I.[CardId] AS CHAR(36)) [CardId],
							I.[Date],
							I.[Total],
							I.[CreatedAt],
							I.[UpdatedAt],
							I.[Active]
						FROM
							dbo.[Invoice] (NOLOCK) I
							INNER JOIN dbo.[Card] (NOLOCK) C ON C.Id = I.CardId
						WHERE
							I.[CardId] = @cardId
							AND C.[UserId] = @userId
						ORDER BY
							[Date]`
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
						AND CONVERT(DATETIME, @endDate)
						ORDER BY [Date]`
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
							@active,
							@invoiceControl
							)`
	GetInvoiceItemByInvoiceId = `SELECT
									CAST(II.[Id] AS CHAR(36)) [Id],
									CAST(II.[InvoiceId] AS CHAR(36)) [InvoiceId],
									CAST(II.[CategoryId] AS CHAR(36)) [CategoryId],
									II.[PurchaseDate],
									II.[Description],
									II.[TotalAmount],
									II.[Installment],
									II.[InstallmentValue],
									II.[Tags],
									II.[InvoiceControl],
									II.[CreatedAt],
									II.[UpdatedAt],
									II.[Active],
									CAST(C.[Id] AS CHAR(36)) [Id],
									C.[Name],
									C.[Active]
								FROM
									dbo.[InvoiceItem] (NOLOCK) II
									INNER JOIN dbo.[Category] (NOLOCK) C ON C.Id = II.CategoryId
									INNER JOIN dbo.[Invoice] (NOLOCK) I ON II.InvoiceId = I.Id
									INNER JOIN dbo.[Card] (NOLOCK) CA ON CA.Id = I.CardId
								WHERE
									II.InvoiceId = @invoiceId
									AND I.CardId = @cardId
									AND CA.UserId = @userId
								ORDER BY
									II.PurchaseDate`
	GetLastInvoiceControl = `SELECT TOP 1 [InvoiceControl] FROM dbo.[InvoiceItem] ORDER BY [InvoiceControl] DESC`
	GetInvoicesCategories = `SELECT
									CAST(II.[InvoiceId] AS CHAR(36)) InvoiceId,
									I.[Date] Date,
									CAST(II.[CategoryId] AS CHAR(36)) CategoryId,
									C.[Name] Name,
									SUM(II.[InstallmentValue]) Total
								FROM
									dbo.[InvoiceItem] II (NOLOCK)
									INNER JOIN dbo.Category C (NOLOCK) ON C.Id = II.CategoryId
									INNER JOIN dbo.[Invoice] I (NOLOCK) ON I.Id = II.InvoiceId
								WHERE
									I.[Date] BETWEEN CONVERT(DATETIME, @startDate)
									AND CONVERT(DATETIME, @endDate)
									AND I.[CardId] = @cardId
								GROUP BY
									II.[CategoryId],
									C.[Name],
									II.[InvoiceId],
									I.[Date]
								ORDER BY
									I.[Date]
								`
	GetInvoiceItemById = `SELECT
							CAST(II.[Id] AS CHAR(36)) [Id],
							CAST(II.[InvoiceId] AS CHAR(36)) [InvoiceId],
							CAST(II.[CategoryId] AS CHAR(36)) [CategoryId],
							II.[PurchaseDate],
							II.[Description],
							II.[TotalAmount],
							II.[Installment],
							II.[InstallmentValue],
							II.[Tags],
							II.[InvoiceControl],
							II.[CreatedAt],
							II.[UpdatedAt],
							II.[Active]
						FROM
							dbo.[InvoiceItem] (NOLOCK) II
						WHERE
							II.Id = @id`
	GetInvoiceById = `SELECT
							CAST(I.[Id] AS CHAR(36)) [Id],
							I.[Date],
							I.[Total],
							I.[CreatedAt],
							I.[UpdatedAt],
							I.[Active],
							CAST(C.[Id] AS CHAR(36)) [CardId],
							C.[Name],
							CAST(II.[Id] AS CHAR(36)) [Id],
							CAST(II.[InvoiceId] AS CHAR(36)) [InvoiceId],
							CAST(II.[CategoryId] AS CHAR(36)) [CategoryId],
							II.[PurchaseDate],
							II.[Description],
							II.[TotalAmount],
							II.[Installment],
							II.[InstallmentValue],
							II.[Tags],
							II.[InvoiceControl],
							II.[CreatedAt],
							II.[UpdatedAt],
							II.[Active],
							CAST(CA.[Id] AS CHAR(36)) [CategoryId],
							CA.[Name],
							CA.[Active]
						FROM
							dbo.Invoice I (NOLOCK)
							INNER JOIN dbo.InvoiceItem II (NOLOCK) ON I.Id = II.InvoiceId
							INNER JOIN dbo.Card C (NOLOCK) ON I.CardId = C.Id
							INNER JOIN dbo.Category CA (NOLOCK) ON II.CategoryId = CA.Id
						WHERE
							I.Id = @id 
						ORDER BY II.[PurchaseDate]`
	GetInvoiceItemByInvoiceControl = `SELECT
										CAST(II.[Id] AS CHAR(36)) [Id],
										CAST(II.[InvoiceId] AS CHAR(36)) [InvoiceId],
										CAST(II.[CategoryId] AS CHAR(36)) [CategoryId],
										II.PurchaseDate,
										II.Description,
										II.TotalAmount,
										II.Installment,
										II.InstallmentValue,
										II.Tags,
										II.CreatedAt,
										II.UpdatedAt,
										II.Active,
										II.InvoiceControl
									FROM
										dbo.InvoiceItem II (NOLOCK)
									WHERE
										II.InvoiceControl = @invoiceControl
									ORDER BY
										II.InvoiceControl`
)
