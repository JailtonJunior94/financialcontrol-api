package queries

const (
	AddUser = `INSERT INTO dbo.[User] VALUES (@id, @name, @email, @password, @createdAt, @updatedAt, @active)`
)
