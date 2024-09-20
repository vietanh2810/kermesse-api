package e2e

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/request"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/response"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/config"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/pkg/dockertester"
)

var (
	testSignupReq = request.SignupRequest{
		Email:           "000@test.com",
		Password:        "hello@123",
		ConfirmPassword: "hello@123",
	}
	testLoginReq = request.SignupRequest{
		Email:    "123@test.com",
		Password: "hello@123",
	}
	testDBErr = errors.New("DB error")
)

type AuthHandlerTestSuite struct {
	suite.Suite

	db       *gorm.DB
	pool     *dockertest.Pool
	resource *dockertest.Resource
	server   *api.Server
}

func (s *AuthHandlerTestSuite) SetupSuite() {
	// Initialize container.
	dt := dockertester.InitPostgres()
	s.pool = dt.Pool
	s.resource = dt.Resource

	// Open connection.
	db, err := dockertester.OpenPostgres(dt.Resource, dt.HostPort)
	require.NoError(s.T(), err)

	s.db = db
}

func (s *AuthHandlerTestSuite) TearDownSuite() {
	err := s.pool.Purge(s.resource) // Destroy the container.
	require.NoError(s.T(), err)
}

func (s *AuthHandlerTestSuite) SetupTest() {
	// Run migrations.
	err := dao.InitTables(s.db)
	require.NoError(s.T(), err)

	// Seed database.
	script, err := os.ReadFile("../scripts/seed_db.sql")
	require.NoError(s.T(), err)

	err = s.db.Exec(string(script)).Error
	require.NoError(s.T(), err)

	// Create API server.
	s.server = api.NewServer(&config.AppConfig{
		API: &config.APIConfig{},
		Gin: &config.GinConfig{
			Mode: gin.TestMode,
		},
		Postgres: &config.PostgresConfig{},
	}, s.db)
}

func (s *AuthHandlerTestSuite) TearDownTest() {
	s.cleanDB()
}

func (s *AuthHandlerTestSuite) cleanDB() {
	script, err := os.ReadFile("../scripts/clean_db.sql")
	require.NoError(s.T(), err)

	err = s.db.Exec(string(script)).Error
	require.NoError(s.T(), err)
}

func (s *AuthHandlerTestSuite) createDBError() {
	// Create/fake a DB error by dropping the users table.
	err := s.db.Exec(`DROP TABLE "users"`).Error
	require.NoError(s.T(), err)
}

func TestAuthHandler(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (s *AuthHandlerTestSuite) TestAuthHandler_HandleSignup() {
	type args struct {
		buildReqBody func() string
	}
	type want struct {
		user     domain.User
		respCode int
		err      *response.Err
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    want
		wantErr bool
	}{
		{
			name:  "201 - Created",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(testSignupReq)
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user: domain.User{
					Email: "000@test.com",
				},
				respCode: http.StatusCreated,
				err:      nil,
			},
			wantErr: false,
		},
		{
			name:  "400 Bad Request - Missing fields",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.SignupRequest{})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("confirm_password: cannot be blank; email: cannot be blank; password: cannot be blank.")),
			},
			wantErr: true,
		},
		{
			name:  "400 Bad Request - Invalid password format",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.SignupRequest{
						Email:           "000@test.com",
						Password:        "Test",
						ConfirmPassword: "Test",
					})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("the password must be at least 8 characters and contain 1 letter, 1 number and 1 symbol")),
			},
			wantErr: true,
		},
		{
			name:  "400 Bad Request - Passwords mismatch",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.SignupRequest{
						Email:           "000@test.com",
						Password:        "Test!123",
						ConfirmPassword: "Test!456",
					})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("confirm password doesn't match the password")),
			},
			wantErr: true,
		},
		{
			name:  "400 Bad Request - User email exists",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.SignupRequest{
						Email:           "123@test.com",
						Password:        "Test!123",
						ConfirmPassword: "Test!123",
					})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("user already exists")),
			},
			wantErr: true,
		},
		{
			name:  "400 Bad Request - invalid JSON",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					return "["
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("unexpected EOF")),
			},
			wantErr: true,
		},
		{
			name: "500 - DB error",
			setup: func() {
				s.createDBError()
			},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(testSignupReq)
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusInternalServerError,
				err:      response.ErrInternalServerError(testDBErr),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			// Setup tests when present.
			tt.setup()

			// Prepare Request.
			body := tt.args.buildReqBody()
			req, err := http.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(body))
			require.NoError(t, err)

			// Execute Request.
			resp := executeRequest(req, s.server)

			// Check the response code.
			assert.Equal(t, tt.want.respCode, resp.Code)

			if tt.wantErr {
				var result response.Err
				err := json.Unmarshal(resp.Body.Bytes(), &result)

				assert.NoError(t, err)
				assert.Equal(t, tt.want.err.ErrorMsg, result.ErrorMsg)
				assert.Equal(t, tt.want.err.ErrorCode, result.ErrorCode)
			} else {
				var result domain.User
				err := json.Unmarshal(resp.Body.Bytes(), &result)

				assert.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, tt.want.user.Email, result.Email)
			}
		})
	}
}

func (s *AuthHandlerTestSuite) TestAuthHandler_HandleLogin() {
	type args struct {
		buildReqBody func() string
	}
	type want struct {
		user     domain.User
		respCode int
		err      *response.Err
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    want
		wantErr bool
	}{
		{
			name:  "200 - OK",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(testLoginReq)
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user: domain.User{
					ID:       123,
					Email:    "123@test.com",
					Password: "Test@123",
				},
				respCode: http.StatusOK,
				err:      nil,
			},
			wantErr: false,
		},
		{
			name:  "400 Bad Request - Missing fields",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.LoginRequest{})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("email: cannot be blank; password: cannot be blank.")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - Wrong password",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.LoginRequest{
						Email:    "000@test.com",
						Password: "Test",
					})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrWrongCredentials(errors.New("wrong password")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - User not found",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(request.LoginRequest{
						Email:    "000@test.com",
						Password: "Test",
					})
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrWrongCredentials(errors.New("user not found")),
			},
			wantErr: true,
		},
		{
			name:  "400 Bad Request - invalid JSON",
			setup: func() {},
			args: args{
				buildReqBody: func() string {
					return "["
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrBadRequest(errors.New("unexpected EOF")),
			},
			wantErr: true,
		},
		{
			name: "500 - DB error",
			setup: func() {
				s.createDBError()
			},
			args: args{
				buildReqBody: func() string {
					body, err := json.Marshal(testSignupReq)
					require.NoError(s.T(), err)

					return string(body)
				},
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusInternalServerError,
				err:      response.ErrInternalServerError(testDBErr),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			// Setup tests when present.
			tt.setup()

			// Prepare Request.
			body := tt.args.buildReqBody()
			req, err := http.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
			require.NoError(t, err)

			// Execute Request.
			resp := executeRequest(req, s.server)

			// Check the response code.
			assert.Equal(t, tt.want.respCode, resp.Code)

			if tt.wantErr {
				var result response.Err
				err := json.Unmarshal(resp.Body.Bytes(), &result)

				assert.NoError(t, err)
				assert.Equal(t, tt.want.err.ErrorMsg, result.ErrorMsg)
				assert.Equal(t, tt.want.err.ErrorCode, result.ErrorCode)
			} else {
				var result response.LoginResponse
				err := json.Unmarshal(resp.Body.Bytes(), &result)

				assert.NoError(t, err)
				assert.EqualValues(t, tt.want.user.ID, result.User.ID)
				assert.Equal(t, tt.want.user.Email, result.User.Email)
				assert.NotEmpty(t, result.Token)
			}
		})
	}
}
