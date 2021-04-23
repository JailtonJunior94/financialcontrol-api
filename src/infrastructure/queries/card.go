package queries

const (
	GetCards = `SELECT
					CAST(C.[Id] AS CHAR(36)) [Id],
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
					CAST(F.[Id] AS CHAR(36)) [FlagId],
					F.[Name],
					F.[Active]
				FROM
					dbo.[Card] C (NOLOCK)
					INNER JOIN dbo.[Flag] F (NOLOCK) ON F.[Id] = C.[FlagId]
				WHERE C.[Active] = 1
				AND C.[UserId] = @userId`
	GetCardById = `SELECT
						CAST(C.[Id] AS CHAR(36)) [Id],
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
						CAST(F.[Id] AS CHAR(36)) [FlagId],
						F.[Name],
						F.[Active]
					FROM
						dbo.[Card] C (NOLOCK)
						INNER JOIN dbo.[Flag] F (NOLOCK) ON F.[Id] = C.[FlagId]
					WHERE C.[Active] = 1
					AND C.[UserId] = @userId
					AND C.[Id] = @id`
	AddCard = `INSERT INTO dbo.[Card]
				VALUES
					(
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
	UpdateCard = `UPDATE dbo.[Card]
				  	SET
						[FlagId] = @flagId,
						[Name] = @name,
						[Number] = @number,
						[Description] = @description,
						[ClosingDay] = @closingDay,
						[ExpirationDate] = @expirationDate,
						[UpdatedAt] = @updatedAt,
						[Active] = @active
				  	WHERE [Id] = @id`
)
