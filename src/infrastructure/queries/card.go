package queries

const (
	GetCards = `SELECT
					CAST([Id] AS CHAR(36)) [Id],
					CAST([UserId] AS CHAR(36)) [UserId],
					CAST([FlagId] AS CHAR(36)) [FlagId],
					[Name],
					[Number],
					[Description],
					[ClosingDay],
					[ExpirationDate],
					[CreatedAt],
					[UpdatedAt],
					[Active]
				FROM dbo.[Card] (NOLOCK)
				WHERE [Active] = 1
				AND [UserId] = @userId`
	GetCardById = `SELECT
						CAST([Id] AS CHAR(36)) [Id],
						CAST([UserId] AS CHAR(36)) [UserId],
						CAST([FlagId] AS CHAR(36)) [FlagId],
						[Name],
						[Number],
						[Description],
						[ClosingDay],
						[ExpirationDate],
						[CreatedAt],
						[UpdatedAt],
						[Active]
					FROM dbo.[Card] (NOLOCK)
					WHERE [Id] = @id
					AND [Active] = 1
					AND [UserId] = @userId`
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
