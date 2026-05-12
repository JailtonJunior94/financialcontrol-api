package mssql

const (
	categorySelectColumns = `
		CAST([Id] AS CHAR(36)) [Id],
		[Name],
		[Sequence],
		[CreatedAt],
		[UpdatedAt],
		[Active]`

	getCategoryByID = `SELECT` + categorySelectColumns + `
		FROM dbo.[Category] (NOLOCK)
		WHERE [Id] = @id
		  AND [Active] = 1`

	getCategoryByIDIncludingDeleted = `SELECT` + categorySelectColumns + `
		FROM dbo.[Category] (NOLOCK)
		WHERE [Id] = @id`

	getActiveChildren = `SELECT` + categorySelectColumns + `
		FROM dbo.[Category] (NOLOCK)
		WHERE 1 = 0`

	addCategory = `INSERT INTO dbo.[Category]
		([Id], [Name], [Sequence], [CreatedAt], [UpdatedAt], [Active])
		VALUES (@id, @name, @sequence, @createdAt, @updatedAt, @active)`

	updateCategory = `UPDATE dbo.[Category]
		SET [Name]      = @name,
		    [Sequence]  = @sequence,
		    [UpdatedAt] = @updatedAt,
		    [Active]    = @active
		WHERE [Id]     = @id
		  AND [Active] = 1`

	softDeleteCascade = `UPDATE dbo.[Category]
		SET [Active]    = 0,
		    [UpdatedAt] = @updatedAt
		WHERE [Id]      = @rootId
		  AND [Active]  = 1`

	categoryExistsForUser = `SELECT COUNT(1)
		FROM dbo.[Category] (NOLOCK)
		WHERE [Id] = @id`

	existsByNameRoot = `SELECT COUNT(1)
		FROM dbo.[Category] (NOLOCK)
		WHERE [Name] = @name
		  AND [Active] = 1
		  AND (@excludeId IS NULL OR CAST([Id] AS CHAR(36)) <> @excludeId)`

	listCategoriesBase = `FROM dbo.[Category] (NOLOCK)
		WHERE [Active] = 1
		  AND (@nameLike IS NULL OR [Name] LIKE @nameLike)`

	listCategoriesCount = `SELECT COUNT(1) ` + listCategoriesBase

	listCategories = `SELECT` + categorySelectColumns + ` ` + listCategoriesBase + `
		ORDER BY [Sequence] ASC, [CreatedAt] DESC`

	listCategoriesPaginated = listCategories + `
		OFFSET @offset ROWS FETCH NEXT @size ROWS ONLY`
)
