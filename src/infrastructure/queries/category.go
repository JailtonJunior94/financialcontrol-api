package queries

const (
	GetCategories = `SELECT
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
