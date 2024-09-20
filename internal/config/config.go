package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
)

type AppConfig struct {
	API      *APIConfig      `mapstructure:"API"`
	Gin      *GinConfig      `mapstructure:"GIN"`
	Postgres *PostgresConfig `mapstructure:"POSTGRES"`
}

func (c *AppConfig) validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.API, validation.Required),
		validation.Field(&c.Gin, validation.Required),
		validation.Field(&c.Postgres, validation.Required),
	)
}

func (c *AppConfig) validateConfig() error {
	if err := c.validate(); err != nil {
		return fmt.Errorf("c.validate() -> %w", err)
	}

	if err := c.API.validate(); err != nil {
		return fmt.Errorf("c.API.validate() -> %w", err)
	}

	if err := c.Gin.validate(); err != nil {
		return fmt.Errorf("c.Gin.validate() -> %w", err)
	}

	if err := c.Postgres.validate(); err != nil {
		return fmt.Errorf("c.Postgres.validate() -> %w", err)
	}

	return nil
}

type APIConfig struct {
	Environment        string   `mapstructure:"ENV"`
	Port               string   `mapstructure:"PORT"`
	BaseURL            string   `mapstructure:"BASE_URL"`
	AllowedCORSDomains []string `mapstructure:"ALLOWED_CORS_DOMAINS"`
	JWTSigningKey      string   `mapstructure:"JWT_SIGNING_KEY"`
}

func (c *APIConfig) validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.Environment, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.BaseURL, validation.Required),
		validation.Field(&c.JWTSigningKey, validation.Required),
	)
}

type GinConfig struct {
	Mode string `mapstructure:"MODE"`
}

func (c *GinConfig) validate() error {
	allowedModes := []any{
		gin.TestMode,
		gin.DebugMode,
		gin.ReleaseMode,
	}

	return validation.ValidateStruct(
		c,
		validation.Field(&c.Mode, validation.Required, validation.In(allowedModes...)),
	)
}

type PostgresConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     string `mapstructure:"PORT"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
	DB       string `mapstructure:"DB"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

func (c *PostgresConfig) validate() error {
	return validation.ValidateStruct(
		c,
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.Password, validation.Required),
		validation.Field(&c.DB, validation.Required),
	)
}
