package mssql

const (
	categorySelectColumns = `
		CAST([Id] AS CHAR(36))       [Id],
		CAST([UserId] AS CHAR(36))   [UserId],
		CAST([ParentId] AS CHAR(36)) [ParentId],
		[Name],
		[Color],
		[Icon],
		[CreatedAt],
		[UpdatedAt],
		[DeletedAt]`

	getCategoryByID = `SELECT` + categorySelectColumns + `
		FROM dbo.[Category] (NOLOCK)
		WHERE [UserId] = @userId
		  AND [Id] = @id
		  AND [DeletedAt] IS NULL`

	getCategoryByIDIncludingDeleted = `SELECT` + categorySelectColumns + `
		FROM dbo.[Category] (NOLOCK)
		WHERE [UserId] = @userId
		  AND [Id] = @id`

	getActiveChildren = `SELECT` + categorySelectColumns + `
		FROM dbo.[Category] (NOLOCK)
		WHERE [UserId] = @userId
		  AND [ParentId] = @parentId
		  AND [DeletedAt] IS NULL
		ORDER BY [Name] ASC`

	addCategory = `INSERT INTO dbo.[Category]
		([Id], [UserId], [ParentId], [Name], [Color], [Icon], [CreatedAt], [UpdatedAt], [DeletedAt])
		VALUES (@id, @userId, @parentId, @name, @color, @icon, @createdAt, @updatedAt, @deletedAt)`

	updateCategory = `UPDATE dbo.[Category]
		SET [ParentId]  = @parentId,
		    [Name]      = @name,
		    [Color]     = @color,
		    [Icon]      = @icon,
		    [UpdatedAt] = @updatedAt,
		    [DeletedAt] = @deletedAt
		WHERE [Id]     = @id
		  AND [UserId] = @userId
		  AND [DeletedAt] IS NULL`

	// Single-statement soft delete cascade: matches the root row and any active
	// child whose ParentId equals the root id, scoped to the user. Already
	// soft-deleted rows are filtered out, making the call idempotent.
	softDeleteCascade = `UPDATE dbo.[Category]
		SET [DeletedAt] = @deletedAt,
		    [UpdatedAt] = @deletedAt
		WHERE [UserId] = @userId
		  AND [DeletedAt] IS NULL
		  AND ([Id] = @rootId OR [ParentId] = @rootId)`

	categoryExistsForUser = `SELECT COUNT(1)
		FROM dbo.[Category] (NOLOCK)
		WHERE [Id] = @id AND [UserId] = @userId`

	existsByNameRoot = `SELECT COUNT(1)
		FROM dbo.[Category] (NOLOCK)
		WHERE [UserId] = @userId
		  AND [ParentId] IS NULL
		  AND [Name] = @name
		  AND [DeletedAt] IS NULL
		  AND (@excludeId IS NULL OR CAST([Id] AS CHAR(36)) <> @excludeId)`

	existsByNameSub = `SELECT COUNT(1)
		FROM dbo.[Category] (NOLOCK)
		WHERE [UserId] = @userId
		  AND [ParentId] = @parentId
		  AND [Name] = @name
		  AND [DeletedAt] IS NULL
		  AND (@excludeId IS NULL OR CAST([Id] AS CHAR(36)) <> @excludeId)`

	listCategoriesBase = `FROM dbo.[Category] (NOLOCK)
		WHERE [UserId] = @userId
		  AND [DeletedAt] IS NULL
		  AND (@nameLike IS NULL OR [Name] LIKE @nameLike)
		  AND (@onlyRoots = 0 OR [ParentId] IS NULL)
		  AND (@onlySubs = 0 OR [ParentId] IS NOT NULL)
		  AND (@parentId IS NULL OR [ParentId] = @parentId)`

	listCategoriesCount = `SELECT COUNT(1) ` + listCategoriesBase

	listCategories = `SELECT` + categorySelectColumns + ` ` + listCategoriesBase + `
		ORDER BY [CreatedAt] DESC`

	listCategoriesPaginated = listCategories + `
		OFFSET @offset ROWS FETCH NEXT @size ROWS ONLY`
)
