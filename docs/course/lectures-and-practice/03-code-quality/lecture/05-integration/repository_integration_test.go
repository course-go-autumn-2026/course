//go:build integration

package integrationdemo

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

const defaultDatabaseURL = "postgres://lecture:lecture@localhost:5434/lecture?sslmode=disable"

type RepositoryTestSuite struct {
	suite.Suite
	pool   *pgxpool.Pool
	repo   *Repository
	userID int
}

func (s *RepositoryTestSuite) SetupSuite() {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	var err error
	s.pool, err = pgxpool.New(context.Background(), databaseURL)
	s.Require().NoError(err)
	s.Require().NoError(s.pool.Ping(context.Background()))

	s.repo = NewRepository(s.pool)
}

func (s *RepositoryTestSuite) TearDownSuite() {
	s.pool.Close()
}

func (s *RepositoryTestSuite) SetupTest() {
	err := s.pool.QueryRow(
		context.Background(),
		`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`,
		"John Doe",
		"john@example.com",
	).Scan(&s.userID)
	s.Require().NoError(err)
}

func (s *RepositoryTestSuite) TearDownTest() {
	_, err := s.pool.Exec(
		context.Background(),
		`DELETE FROM users WHERE id = $1`,
		s.userID,
	)
	s.Require().NoError(err)
}

func (s *RepositoryTestSuite) TestGetByID() {
	user, err := s.repo.GetByID(context.Background(), s.userID)

	s.Require().NoError(err)
	s.Equal(s.userID, user.ID)
	s.Equal("John Doe", user.Name)
	s.Equal("john@example.com", user.Email)
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
