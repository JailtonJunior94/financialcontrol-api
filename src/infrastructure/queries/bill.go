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
					[Active] = 1`
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
	AddBill = `INSERT INTO dbo.[Bill] VALUES (@id, @date, @total, @sixtyPercent, @fortyPercent, @createdAt, @updatedAt, @active)`
)
