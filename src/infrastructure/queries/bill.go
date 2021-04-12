package queries

const (
	GetBills = `SELECT
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
	GetBillByDate = `SELECT
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
	GetBillById = `SELECT
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
	AddBill    = `INSERT INTO dbo.[Bill] VALUES (@id, @date, @total, @sixtyPercent, @fortyPercent, @createdAt, @updatedAt, @active)`
	UpdateBill = `UPDATE
						dbo.[Bill]
					SET
						[Total] = @total,
						[SixtyPercent] = @sixtyPercent,
						[FortyPercent] = @fortyPercent,
						[UpdatedAt] = @updatedAt
					WHERE
						[Id] = @id`
	AddBillItem         = `INSERT INTO dbo.[BillItem] VALUES (@id, @billId, @title, @value, @createdAt, @updatedAt, @active)`
	GetBillItemByBillId = `SELECT
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
								AND [BillId] = @billId`
	GetBillItemById = `SELECT
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
)
