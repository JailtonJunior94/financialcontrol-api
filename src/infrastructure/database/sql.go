package database

import (
	"log"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/jmoiron/sqlx"
)

type ISqlConnection interface {
	Connect() *sqlx.DB
	Disconnect()
}

type SqlConnection struct {
	db *sqlx.DB
}

func NewConnection() ISqlConnection {
	db, err := sqlx.Connect("sqlserver", environments.SqlConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return &SqlConnection{db}
}

func (s *SqlConnection) Connect() *sqlx.DB {
	return s.db
}

func (s *SqlConnection) Disconnect() {
	if err := s.db.Close(); err != nil {
		log.Fatal(err)
	}
}
