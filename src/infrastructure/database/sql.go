package database

import (
	"database/sql"
	"log"

	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/jmoiron/sqlx"
)

type ISqlConnection interface {
	Connect() *sqlx.DB
	Disconnect()
	OpenConnectionAndMountStatement(query string) (*sql.Stmt, error)
	ValidateResult(result sql.Result, err error) error
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

func (s *SqlConnection) OpenConnectionAndMountStatement(query string) (*sql.Stmt, error) {
	stmt, err := s.db.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

func (s *SqlConnection) ValidateResult(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if rows == 0 {
		return err
	}
	return nil
}
