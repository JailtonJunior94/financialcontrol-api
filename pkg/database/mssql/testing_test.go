//go:build integration

package mssql_test

import (
	"context"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
	"github.com/stretchr/testify/suite"
)

type MSSQLTestDatabaseSuite struct {
	suite.Suite
}

func TestMSSQLTestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(MSSQLTestDatabaseSuite))
}

func (s *MSSQLTestDatabaseSuite) TestGetSharedTestDatabase() {
	scenarios := []struct {
		name   string
		setup  func()
		expect func()
	}{
		{
			name:  "deve conectar ao container e executar SELECT 1 com sucesso",
			setup: func() {},
			expect: func() {
				db, cleanup, err := mssql.GetSharedTestDatabase()
				s.Require().NoError(err, "GetSharedTestDatabase não deve retornar erro")
				s.Require().NotNil(db, "db não pode ser nil")
				s.Require().NotNil(cleanup, "cleanup não pode ser nil")
				defer cleanup()

				var result int
				err = db.QueryRowContext(context.Background(), "SELECT 1").Scan(&result)
				s.Require().NoError(err, "SELECT 1 deve funcionar")
				s.Equal(1, result)
			},
		},
		{
			name:  "singleton: segunda chamada retorna o mesmo db sem erro",
			setup: func() {},
			expect: func() {
				db1, cleanup1, err1 := mssql.GetSharedTestDatabase()
				s.Require().NoError(err1)
				defer cleanup1()

				db2, cleanup2, err2 := mssql.GetSharedTestDatabase()
				s.Require().NoError(err2)
				defer cleanup2()

				s.Same(db1, db2, "singleton deve retornar a mesma instância de *sqlx.DB")
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			sc.expect()
		})
	}
}
