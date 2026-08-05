package main

import (
	"fmt"
	"time"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/spf13/cobra"
)

type Config struct {
	DSN     core.Secret        `mapstructure:"dsn" validate:"required"`
	HTTP    HTTPConfig         `mapstructure:"http" validate:"required"`
	Storage core.StorageConfig `mapstructure:"storage"`
	Auth    AuthConfig         `mapstructure:"auth" validate:"required"`
	Email   EmailConfig        `mapstructure:"email"`
}

type HTTPConfig struct {
	Addr                string        `mapstructure:"addr" validate:"required"`
	ReadTimeout         time.Duration `mapstructure:"read_timeout" validate:"gt=0"`
	WriteTimeout        time.Duration `mapstructure:"write_timeout" validate:"gt=0"`
	IdleTimeout         time.Duration `mapstructure:"idle_timeout" validate:"gt=0"`
	ShutdownTimeout     time.Duration `mapstructure:"shutdown_timeout" validate:"gt=0"`
	MaxHeaderBytes      int           `mapstructure:"max_header_bytes" validate:"gt=0"`
	MaxRequestBodyBytes int64         `mapstructure:"max_request_body_bytes" validate:"gt=0"`
}

type AuthConfig struct {
	JWTSecret            core.Secret   `mapstructure:"jwt_secret" validate:"required,min=32"`
	AccessTokenLifetime  time.Duration `mapstructure:"access_token_lifetime" validate:"gt=0"`
	RefreshTokenLifetime time.Duration `mapstructure:"refresh_token_lifetime" validate:"gt=0"`
	ResetTokenLifetime   time.Duration `mapstructure:"reset_token_lifetime" validate:"gt=0"`
	PasswordResetURL     string        `mapstructure:"password_reset_url"`
}

type EmailConfig struct {
	SMTP SMTPConfig `mapstructure:"smtp"`
}

type SMTPConfig struct {
	Host     string      `mapstructure:"host"`
	Port     int         `mapstructure:"port"`
	Username string      `mapstructure:"username"`
	Password core.Secret `mapstructure:"password"`
	From     string      `mapstructure:"from"`
}

// loadConfig builds a Config by merging:
//
// flags > environment > config file > defaults
func loadConfig(cmd *cobra.Command, cfgFile string) (Config, error) {
	v := core.NewViper()

	v.SetDefault("dsn", "")
	v.SetDefault("auth.jwt_secret", "")
	v.SetDefault("auth.access_token_lifetime", 15*time.Minute)
	v.SetDefault("auth.refresh_token_lifetime", 30*24*time.Hour)
	v.SetDefault("auth.reset_token_lifetime", time.Hour)
	v.SetDefault("email.smtp.port", 587)
	v.SetDefault("http.addr", ":8080")
	v.SetDefault("http.read_timeout", 10*time.Second)
	v.SetDefault("http.write_timeout", 10*time.Second)
	v.SetDefault("http.idle_timeout", 60*time.Second)
	v.SetDefault("http.shutdown_timeout", 10*time.Second)
	v.SetDefault("http.max_header_bytes", 1<<20)
	v.SetDefault("http.max_request_body_bytes", 1<<20)

	if err := core.ReadConfigFile(v, cfgFile); err != nil {
		return Config{}, err
	}

	// Core flags.
	bindFlag := func(key, name string) {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			_ = v.BindPFlag(key, flag)
		}
	}
	bindFlag("http.addr", "addr")
	bindFlag("dsn", "dsn")
	bindFlag("auth.jwt_secret", "jwt-secret")

	// Storage flags.
	core.BindStorageFlags(v, cmd)

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}
