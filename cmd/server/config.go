package main

import (
	"fmt"
	"strings"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
	"github.com/UnivocalX/odessa/internal/storage/azure"
	"github.com/UnivocalX/odessa/internal/storage/fs"
	"github.com/UnivocalX/odessa/internal/storage/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config mirrors the viper key space.
//
// Sources (highest → lowest priority):
//
//  1. CLI flags
//     --addr
//     --dsn
//     --jwt-secret
//     --fs-root
//     --s3-region
//     --s3-endpoint
//     --az-account
//
//  2. Environment variables (prefix ODESSA_)
//     ODESSA_ADDR
//     ODESSA_DSN
//     ODESSA_AUTH_JWT_SECRET
//
//     ODESSA_STORAGE_FS_ROOT
//
//     ODESSA_STORAGE_S3_REGION
//     ODESSA_STORAGE_S3_ENDPOINT
//
//     ODESSA_STORAGE_AZURE_ACCOUNT
//     ODESSA_STORAGE_AZURE_CONNECTION_STRING
//     ODESSA_STORAGE_AZURE_ACCOUNT_KEY
//     ODESSA_STORAGE_AZURE_USE_DEFAULT_CREDENTIAL
//
//  3. Config file (--config)
//
//     addr: ":8080"
//     dsn: "postgres://..."
//
//     auth:
//     jwt_secret: "super-secret"
//
//     storage:
//     fs:
//     root: "/data"
//
//     s3:
//     region: "us-east-1"
//     endpoint: ""
//
//     azure:
//     account: "storageaccount"
//     connection_string: ""
//     account_key: ""
//     use_default_credential: false
//
//  4. Built-in defaults
//
// A storage backend is enabled when its primary configuration exists and
// disabled is not true.
type Config struct {
	Addr    string            `mapstructure:"addr" validate:"required"`
	DSN     repository.Secret `mapstructure:"dsn" validate:"required"`
	Storage StorageConfig     `mapstructure:"storage"`
	Auth    AuthConfig        `mapstructure:"auth" validate:"required"`
}

type AuthConfig struct {
	JWTSecret repository.Secret `mapstructure:"jwt_secret" validate:"required,min=32"`
}

type StorageConfig struct {
	FS    FSConfig    `mapstructure:"fs"`
	S3    S3Config    `mapstructure:"s3"`
	Azure AzureConfig `mapstructure:"azure"`
}

type FSConfig struct {
	Disabled bool   `mapstructure:"disabled"`
	Root     string `mapstructure:"root"`
}

type S3Config struct {
	Disabled bool   `mapstructure:"disabled"`
	Region   string `mapstructure:"region"`
	Endpoint string `mapstructure:"endpoint"`
}

type AzureConfig struct {
	Disabled             bool   `mapstructure:"disabled"`
	Account              string `mapstructure:"account"`
	ConnectionString     string `mapstructure:"connection_string"`
	AccountKey           string `mapstructure:"account_key"`
	UseDefaultCredential bool   `mapstructure:"use_default_credential"`
}

// loadConfig builds a Config by merging:
//
// flags > environment > config file > defaults
func loadConfig(cmd *cobra.Command, cfgFile string) (Config, error) {
	v := viper.New()

	v.SetEnvPrefix("ODESSA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	v.SetDefault("addr", ":8080")
	v.SetDefault("dsn", "")
	v.SetDefault("auth.jwt_secret", "")

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)

		if err := v.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config %q: %w", cfgFile, err)
		}
	}

	// Core flags.
	_ = v.BindPFlag("addr", cmd.Flags().Lookup("addr"))
	_ = v.BindPFlag("dsn", cmd.Flags().Lookup("dsn"))
	_ = v.BindPFlag("auth.jwt_secret", cmd.Flags().Lookup("jwt-secret"))

	// Storage flags.
	_ = v.BindPFlag("storage.fs.root", cmd.Flags().Lookup("fs-root"))
	_ = v.BindPFlag("storage.s3.region", cmd.Flags().Lookup("s3-region"))
	_ = v.BindPFlag("storage.s3.endpoint", cmd.Flags().Lookup("s3-endpoint"))
	_ = v.BindPFlag("storage.azure.account", cmd.Flags().Lookup("az-account"))

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// configureStorage configures the auto-registered storage backends from config.
func configureStorage(cfg StorageConfig) error {
	// Filesystem
	if cfg.FS.Disabled {
		storage.Unregister("file")
	} else if cfg.FS.Root != "" {
		if err := fs.Configure(fs.WithRoot(cfg.FS.Root)); err != nil {
			return fmt.Errorf("configure fs: %w", err)
		}
	}

	// S3
	if cfg.S3.Disabled {
		storage.Unregister("s3")
	} else if cfg.S3.Region != "" || cfg.S3.Endpoint != "" {
		var opts []s3.Option

		if cfg.S3.Endpoint != "" {
			opts = append(opts, s3.WithEndpoint(cfg.S3.Endpoint, cfg.S3.Region))
		} else {
			opts = append(opts, s3.WithRegion(cfg.S3.Region))
		}

		if err := s3.Configure(opts...); err != nil {
			return fmt.Errorf("configure s3: %w", err)
		}
	}

	// Azure
	if cfg.Azure.Disabled {
		storage.Unregister("azure")
	} else if cfg.Azure.Account != "" {
		var opts []azure.Option

		switch {
		case cfg.Azure.ConnectionString != "":
			opts = append(opts, azure.WithConnectionString(cfg.Azure.ConnectionString))

		case cfg.Azure.AccountKey != "":
			opts = append(opts, azure.WithAccountKey(
				cfg.Azure.Account,
				cfg.Azure.AccountKey,
			))

		default:
			opts = append(opts, azure.WithDefaultCredential(
				cfg.Azure.Account,
			))
		}

		if err := azure.Configure(opts...); err != nil {
			return fmt.Errorf("configure azure: %w", err)
		}
	}

	return nil
}
