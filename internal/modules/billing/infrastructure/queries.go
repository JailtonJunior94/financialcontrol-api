package infrastructure

const (
	getBills = `SELECT
					CAST([Id] AS CHAR(36)) [Id],
					[Date],
					[Total],
					[SixtyPercent],
					[FortyPercent],
					[CreatedAt],
					[UpdatedAt],
					[Active]
				FROM
					dbo.[Bill] (NOLOCK)
				WHERE
					[Active] = 1
				ORDER BY [Date]`
	getBillByDate = `SELECT
						CAST([Id] AS CHAR(36)) [Id],
						[Date],
						[Total],
						[SixtyPercent],
						[FortyPercent],
						[CreatedAt],
						[UpdatedAt],
						[Active]
					FROM
						dbo.[Bill] (NOLOCK)
					WHERE
						[Active] = 1
						AND [Date] BETWEEN CONVERT(DATETIME, @startDate)
						AND CONVERT(DATETIME, @endDate)`
	getBillByID = `SELECT
						CAST([Id] AS CHAR(36)) [Id],
						[Date],
						[Total],
						[SixtyPercent],
						[FortyPercent],
						[CreatedAt],
						[UpdatedAt],
						[Active]
					FROM
						dbo.[Bill] (NOLOCK)
					WHERE
						[Active] = 1
						AND [Id] = @id`
	addBill    = `INSERT INTO dbo.[Bill] VALUES (@id, @date, @total, @sixtyPercent, @fortyPercent, @createdAt, @updatedAt, @active)`
	updateBill = `UPDATE
						dbo.[Bill]
					SET
						[Total] = @total,
						[SixtyPercent] = @sixtyPercent,
						[FortyPercent] = @fortyPercent,
						[UpdatedAt] = @updatedAt
					WHERE
						[Id] = @id`
	getBillItemByBillID = `SELECT
								CAST([Id] AS CHAR(36)) [Id],
								CAST([BillId] AS CHAR(36)) [BillId],
								[Title],
								[Value],
								[CreatedAt],
								[UpdatedAt],
								[Active]
							FROM
								dbo.[BillItem] (NOLOCK)
							WHERE
								[Active] = 1
								AND [BillId] = @billId
							ORDER BY [Value] DESC`
	getBillItemByID = `SELECT
							CAST([Id] AS CHAR(36)) [Id],
							CAST([BillId] AS CHAR(36)) [BillId],
							[Title],
							[Value],
							[CreatedAt],
							[UpdatedAt],
							[Active]
						FROM
							dbo.[BillItem] (NOLOCK)
						WHERE
							[Active] = 1
						AND [Id] = @id
						AND [BillId] = @billId`
	addBillItem    = `INSERT INTO dbo.[BillItem] VALUES (@id, @billId, @title, @value, @createdAt, @updatedAt, @active)`
	updateBillItem = `UPDATE
							dbo.[BillItem]
						SET
							[Title] = @title,
							[Value] = @value,
							[UpdatedAt] = @updatedAt,
							[Active] = @active
						WHERE
							[Id] = @id
						AND [BillId] = @billId`
)
