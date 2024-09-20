package config

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	apiENV                = "test"
	apiPort               = "1234"
	apiBaseURL            = "localhost:" + apiPort
	apiAllowedCORSDomains = "my-domain1.com,my-domain2.com"
	apiJWTSigningKey      = "test_jwt_key"

	ginMode = "debug"

	postgresHost     = "pg"
	postgresPort     = "5678"
	postgresUsername = "root"
	postgresPassword = "pass123"
	postgresDB       = "testDB"
	postgresLogLevel = "error"
)

func TestLoad(t *testing.T) {
	type args struct {
		configFile string
	}
	tests := []struct {
		name       string
		setupENV   func()
		args       args
		want       *AppConfig
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "Happy Path",
			setupENV: func() {
				setENVs(t)
			},
			args: args{
				configFile: "testdata/good.yml",
			},
			want: &AppConfig{
				API: &APIConfig{
					Environment:        apiENV,
					Port:               apiPort,
					BaseURL:            apiBaseURL,
					AllowedCORSDomains: strings.Split(apiAllowedCORSDomains, ","),
					JWTSigningKey:      apiJWTSigningKey,
				},
				Gin: &GinConfig{
					Mode: ginMode,
				},
				Postgres: &PostgresConfig{
					Host:     postgresHost,
					Port:     postgresPort,
					User:     postgresUsername,
					Password: postgresPassword,
					DB:       postgresDB,
					LogLevel: postgresLogLevel,
				},
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name:     "Missing keys",
			setupENV: func() {},
			args: args{
				configFile: "testdata/missing_keys.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "conf.validateConfig -> c.validate() -> API: cannot be blank; Gin: cannot be blank; Postgres: cannot be blank.",
		},
		{
			name:     "Invalid YAML file",
			setupENV: func() {},
			args: args{
				configFile: "testdata/invalid_yaml.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "viper.ReadInConfig -> While parsing config: yaml: line 2: mapping values are not allowed in this context",
		},
		{
			name:     "Unable to marshal",
			setupENV: func() {},
			args: args{
				configFile: "testdata/unmarshallable.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "viper.Unmarshal -> 1 error(s) decoding:\n\n* 'API' expected a map, got 'slice'",
		},
		{
			name: "Invalid API configs - missing port",
			setupENV: func() {
				setENVs(t)

				err := os.Unsetenv("API_PORT")
				require.NoError(t, err)
			},
			args: args{
				configFile: "testdata/good.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: `conf.validateConfig -> c.API.validate() -> Port: cannot be blank.`,
		},
		{
			name: "Invalid Gin configs - missing mode",
			setupENV: func() {
				setENVs(t)

				err := os.Unsetenv("GIN_MODE")
				require.NoError(t, err)
			},
			args: args{
				configFile: "testdata/good.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: `conf.validateConfig -> c.validate() -> Gin: cannot be blank.`,
		},
		{
			name: "Invalid Gin configs - invalid value",
			setupENV: func() {
				setENVs(t)

				err := os.Setenv("GIN_MODE", "unknown")
				require.NoError(t, err)
			},
			args: args{
				configFile: "testdata/good.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: `conf.validateConfig -> c.Gin.validate() -> Mode: must be a valid value.`,
		},
		{
			name: "Invalid Postgres configs - missing DB",
			setupENV: func() {
				setENVs(t)

				err := os.Unsetenv("POSTGRES_DB")
				require.NoError(t, err)
			},
			args: args{
				configFile: "testdata/good.yml",
			},
			want:       nil,
			wantErr:    true,
			wantErrMsg: `conf.validateConfig -> c.Postgres.validate() -> DB: cannot be blank.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupENV()
			defer os.Clearenv()

			got, err := Load(tt.args.configFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Actual error messages may have random spaces, so we remove spaces.
			if err != nil && tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("Load() errorMsg = %v, wantErrMsg %v", err.Error(), tt.wantErrMsg)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func setENVs(t *testing.T) {
	m := map[string]string{
		"API_ENV":                  apiENV,
		"API_PORT":                 apiPort,
		"API_BASE_URL":             apiBaseURL,
		"API_ALLOWED_CORS_DOMAINS": apiAllowedCORSDomains,
		"API_JWT_SIGNING_KEY":      apiJWTSigningKey,
		"GIN_MODE":                 ginMode,
		"POSTGRES_HOST":            postgresHost,
		"POSTGRES_PORT":            postgresPort,
		"POSTGRES_USER":            postgresUsername,
		"POSTGRES_PASSWORD":        postgresPassword,
		"POSTGRES_DB":              postgresDB,
		"POSTGRES_LOG_LEVEL":       postgresLogLevel,
	}

	for k, v := range m {
		err := os.Setenv(k, v)
		require.NoError(t, err)
	}
}
