package db

import (
	"context"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/pkg/dockertester"
)

type UserDBTestSuite struct {
	suite.Suite

	db       *gorm.DB
	pool     *dockertest.Pool
	resource *dockertest.Resource

	userDAO *dao.UserDAO
}

func (s *UserDBTestSuite) SetupSuite() {
	// Initialize container.
	dt := dockertester.InitPostgres()
	s.pool = dt.Pool
	s.resource = dt.Resource

	// Open connection.
	db, err := dockertester.OpenPostgres(dt.Resource, dt.HostPort)
	require.NoError(s.T(), err)

	s.db = db
}

func (s *UserDBTestSuite) TearDownSuite() {
	err := s.pool.Purge(s.resource) // Destroy the container.
	require.NoError(s.T(), err)
}

func (s *UserDBTestSuite) SetupTest() {
	// Run migrations.
	err := dao.InitTables(s.db)
	require.NoError(s.T(), err)

	// Seed database.
	script, err := os.ReadFile("../scripts/seed_db.sql")
	require.NoError(s.T(), err)

	err = s.db.Exec(string(script)).Error
	require.NoError(s.T(), err)

	// Initialize DAO.
	s.userDAO = dao.NewUserDAO(s.db)
}

func (s *UserDBTestSuite) TearDownTest() {
	s.cleanDB()
}

func (s *UserDBTestSuite) cleanDB() {
	script, err := os.ReadFile("../scripts/clean_db.sql")
	require.NoError(s.T(), err)

	err = s.db.Exec(string(script)).Error
	require.NoError(s.T(), err)
}

func TestUserDB(t *testing.T) {
	suite.Run(t, new(UserDBTestSuite))
}

func (s *UserDBTestSuite) TestUserDB_FindByID() {
	result, err := s.userDAO.FindByID(context.TODO(), 12345)
	assert.Error(s.T(), gorm.ErrRecordNotFound)

	result, err = s.userDAO.FindByID(context.TODO(), 123)
	assert.NoError(s.T(), err)

	assert.NotNil(s.T(), result)
	assert.EqualValues(s.T(), 123, result.ID)
	assert.Equal(s.T(), "123@test.com", result.Email)
}

func (s *UserDBTestSuite) TestUserDB_FindByEmail() {
	result, err := s.userDAO.FindByEmail(context.TODO(), "non-exist@test.com")
	assert.Error(s.T(), gorm.ErrRecordNotFound)

	result, err = s.userDAO.FindByEmail(context.TODO(), "123@test.com")
	assert.NoError(s.T(), err)

	assert.NotNil(s.T(), result)
	assert.EqualValues(s.T(), 123, result.ID)
	assert.Equal(s.T(), "123@test.com", result.Email)
}

func (s *UserDBTestSuite) TestUserDB_Insert() {
	result, err := s.userDAO.Insert(context.TODO(), dao.User{
		Email:    "123@test.com",
		Password: "any",
	})
	assert.Error(s.T(), dao.ErrUserEmailExists)

	const (
		id       = 789
		email    = "789@test.com"
		password = "Test@789"
	)
	result, err = s.userDAO.Insert(context.TODO(), dao.User{
		ID:       id,
		Email:    email,
		Password: password,
	})
	assert.NoError(s.T(), err)

	assert.NotNil(s.T(), result)
	assert.EqualValues(s.T(), id, result.ID)
	assert.Equal(s.T(), email, result.Email)
	assert.Equal(s.T(), password, result.Password)
}
