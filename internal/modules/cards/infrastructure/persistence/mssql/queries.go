package mssql

const (
	listCards = `SELECT
					CAST(C.[Id] AS CHAR(36))     [Id],
					CAST(C.[UserId] AS CHAR(36)) [UserId],
					CAST(C.[FlagId] AS CHAR(36)) [FlagId],
					C.[Name],
					C.[Number],
					C.[Description],
					C.[ClosingDay],
					C.[ExpirationDate],
					C.[CreatedAt],
					C.[UpdatedAt],
					C.[Active],
					CAST(F.[Id] AS CHAR(36))     [FlagEntityId],
					F.[Name]                     [FlagName],
					F.[Active]                   [FlagActive]
				FROM
					dbo.[Card] C (NOLOCK)
					INNER JOIN dbo.[Flag] F (NOLOCK) ON F.[Id] = C.[FlagId]
				WHERE C.[Active] = 1
				AND C.[UserId] = @userId
				ORDER BY C.[CreatedAt] DESC`

	listCardsPaginated = listCards + `
				OFFSET @offset ROWS FETCH NEXT @size ROWS ONLY`

	getCardByID = `SELECT
					CAST(C.[Id] AS CHAR(36))     [Id],
					CAST(C.[UserId] AS CHAR(36)) [UserId],
					CAST(C.[FlagId] AS CHAR(36)) [FlagId],
					C.[Name],
					C.[Number],
					C.[Description],
					C.[ClosingDay],
					C.[ExpirationDate],
					C.[CreatedAt],
					C.[UpdatedAt],
					C.[Active],
					CAST(F.[Id] AS CHAR(36))     [FlagEntityId],
					F.[Name]                     [FlagName],
					F.[Active]                   [FlagActive]
				FROM
					dbo.[Card] C (NOLOCK)
					INNER JOIN dbo.[Flag] F (NOLOCK) ON F.[Id] = C.[FlagId]
				WHERE C.[Active] = 1
				AND C.[UserId] = @userId
				AND C.[Id] = @id`

	addCard = `INSERT INTO dbo.[Card]
				VALUES (
					@id,
					@userId,
					@flagId,
					@name,
					@number,
					@description,
					@closingDay,
					@expirationDate,
					@createdAt,
					@updatedAt,
					@active
				)`

	updateCard = `UPDATE dbo.[Card]
				  SET
					[FlagId]         = @flagId,
					[Name]           = @name,
					[Number]         = @number,
					[Description]    = @description,
					[ClosingDay]     = @closingDay,
					[ExpirationDate] = @expirationDate,
					[UpdatedAt]      = @updatedAt,
					[Active]         = @active
				  WHERE [Id] = @id
				  AND   [UserId] = @userId`

	listFlags = `SELECT
					CAST([Id] AS CHAR(36)) [Id],
					[Name],
					[Active]
				FROM dbo.[Flag] (NOLOCK)
				WHERE [Active] = 1
				ORDER BY [Name] ASC`

	existsFlag = `SELECT COUNT(1)
				  FROM dbo.[Flag] (NOLOCK)
				  WHERE CAST([Id] AS CHAR(36)) = @id
				  AND   [Active] = 1`
)
