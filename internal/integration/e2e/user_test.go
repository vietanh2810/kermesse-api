package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/api/handler/v1/response"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/config"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/domain"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/pkg/jwthelper"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/internal/repository/dao"
	"github.com/yizeng/gab/gin/gorm/auth-jwt/pkg/dockertester"
)

const (
	jwtSigningKey = "test_key"
)

type UserHandlerTestSuite struct {
	suite.Suite

	db       *gorm.DB
	pool     *dockertest.Pool
	resource *dockertest.Resource
	server   *api.Server
}

func (s *UserHandlerTestSuite) SetupSuite() {
	// Initialize container.
	dt := dockertester.InitPostgres()
	s.pool = dt.Pool
	s.resource = dt.Resource

	// Open connection.
	db, err := dockertester.OpenPostgres(dt.Resource, dt.HostPort)
	require.NoError(s.T(), err)

	s.db = db
}

func (s *UserHandlerTestSuite) TearDownSuite() {
	err := s.pool.Purge(s.resource) // Destroy the container.
	require.NoError(s.T(), err)
}

func (s *UserHandlerTestSuite) SetupTest() {
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
		API: &config.APIConfig{
			JWTSigningKey: jwtSigningKey,
		},
		Gin: &config.GinConfig{
			Mode: gin.TestMode,
		},
		Postgres: &config.PostgresConfig{},
	}, s.db)
}

func (s *UserHandlerTestSuite) TearDownTest() {
	s.cleanDB()
}

func (s *UserHandlerTestSuite) cleanDB() {
	script, err := os.ReadFile("../scripts/clean_db.sql")
	require.NoError(s.T(), err)

	err = s.db.Exec(string(script)).Error
	require.NoError(s.T(), err)
}

func (s *UserHandlerTestSuite) createDBError() {
	// Create/fake a DB error by dropping the users table.
	err := s.db.Exec(`DROP TABLE "users"`).Error
	require.NoError(s.T(), err)
}

func TestUserHandler(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

func (s *UserHandlerTestSuite) TestUserHandler_HandleGetUser() {
	type args struct {
		createHeaders func() map[string]string
		userID        string
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
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 123, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "123",
			},
			want: want{
				user: domain.User{
					Email: "123@test.com",
				},
				respCode: http.StatusOK,
				err:      nil,
			},
			wantErr: false,
		},
		{
			name:  "401 Unauthorized - No JWT",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					return map[string]string{}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - With invalid authorization header (length shorter than 7)",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					return map[string]string{
						"Authorization": "none",
					}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - With invalid authorization header (not started with BEARER)",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					return map[string]string{
						"Authorization": "TOKEN some-valid-token",
					}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - Malformed token",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					return map[string]string{
						"Authorization": "BEARER not-valid-token",
					}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - Token with wrong signature",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					return map[string]string{
						"Authorization": "BEARER eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDY4MDU3MTAsIlVzZXJJRCI6NiwiVXNlckFnZW50IjoiUG9zdG1hblJ1bnRpbWUvNy4zNi4xIn0.v4IHwPkJYKaA4G00r5C2KeyyQhK93VgmNxAKkf7afL2Ybj9Qtfyv7nn9JBg8nTrjwLWwd_3fdsUGhAl6yRDYFw",
					}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - Mismatching user agent",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 123, "other user agent")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "401 Unauthorized - UserID is empty",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 0, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusUnauthorized,
				err:      response.ErrJWTUnverified(errors.New("any message")),
			},
			wantErr: true,
		},
		{
			name:  "400 Bad Request - Invalid userID",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 123, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "abc",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusBadRequest,
				err:      response.ErrInvalidInput("userID", "abc"),
			},
			wantErr: true,
		},
		{
			name:  "404 Not Found - Negative userID",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 123, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "-123",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusNotFound,
				err:      response.ErrNotFound("user", "ID", "-123"),
			},
			wantErr: true,
		},
		{
			name:  "403 Forbidden- JWT UserID doesn't match userID in URL query",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 123, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "456",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusForbidden,
				err:      response.ErrPermissionDenied(fmt.Errorf("can't view user %v by user %v", "456", "123")),
			},
			wantErr: true,
		},
		{
			name:  "404 - User Not Found",
			setup: func() {},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 456, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "456",
			},
			want: want{
				user:     domain.User{},
				respCode: http.StatusNotFound,
				err:      response.ErrNotFound("user", "ID", "456"),
			},
			wantErr: true,
		},
		{
			name: "500 - DB error",
			setup: func() {
				s.createDBError()
			},
			args: args{
				createHeaders: func() map[string]string {
					token, err := jwthelper.GenerateToken([]byte(jwtSigningKey), 123, "")
					require.NoError(s.T(), err)

					return map[string]string{
						"Authorization": "Bearer " + token,
					}
				},
				userID: "123",
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
			req, err := http.NewRequest("GET", "/api/v1/users/"+tt.args.userID, nil)
			require.NoError(t, err)

			headers := tt.args.createHeaders()
			for k, v := range headers {
				req.Header.Set(k, v)
			}

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
