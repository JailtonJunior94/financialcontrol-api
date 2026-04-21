package infrastructure

const (
	getFlags = `SELECT
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

	getCategories = `SELECT
						CAST([Id] AS CHAR(36)) [Id],
						[Name],
						[Sequence],
						[CreatedAt],
						[UpdatedAt],
						[Active]
					FROM
						dbo.[Category] (NOLOCK)
					WHERE
						[Active] = 1
					ORDER BY
						[Sequence]`
)
