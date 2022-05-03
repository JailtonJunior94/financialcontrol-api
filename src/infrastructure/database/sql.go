package database

import (
	"context"
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
	IUnitOfWork
}

type IUnitOfWork interface {
	Begin() (*sqlx.Tx, error)
	Rollback() error
	Commit() error
	End(txFunc func() error) error
}

type SqlConnection struct {
	DB *sqlx.DB
	TX *sqlx.Tx
}

func NewConnection() ISqlConnection {
	db, err := sqlx.Connect("sqlserver", environments.SqlConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return &SqlConnection{DB: db}
}

func (s *SqlConnection) Connect() *sqlx.DB {
	return s.DB
}

func (s *SqlConnection) Disconnect() {
	if err := s.DB.Close(); err != nil {
		log.Fatal(err)
	}
}

func (s *SqlConnection) OpenConnectionAndMountStatement(query string) (*sql.Stmt, error) {
	stmt, err := s.DB.DB.Prepare(query)
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

func (s *SqlConnection) Begin() (*sqlx.Tx, error) {
	tx, err := s.DB.BeginTxx(context.Background(), nil)
	return tx, err
}

func (s *SqlConnection) Rollback() error {
	return s.TX.Rollback()
}

func (s *SqlConnection) Commit() error {
	return s.TX.Rollback()
}

func (s *SqlConnection) End(txFunc func() error) error {
	var err error
	tx := s.TX

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc()
	return err
}
