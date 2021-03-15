package queries

const (
	AddUser    = `INSERT INTO dbo.[User] VALUES (@id, @name, @email, @password, @createdAt, @updatedAt, @active)`
	GetByEmail = `SELECT
					CAST([Id] AS CHAR(36)) [Id],
					[Name],
					[Email],
					[Password],
					[CreatedAt],
					[UpdatedAt],
					[Active]
				FROM
					dbo.[User] (NOLOCK)
				WHERE
					[Email] = @email`
)
