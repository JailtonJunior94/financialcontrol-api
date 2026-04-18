package queries

const (
	GetFlags = `SELECT
					CAST([Id] AS CHAR(36)) [Id],
					[Name],
					[CreatedAt],
					[UpdatedAt],
					[Active]
				FROM
					dbo.[Flag] (NOLOCK)
				WHERE
					[Active] = 1
				ORDER BY
					[Name]`
)
