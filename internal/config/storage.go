package config

import (
	"fmt"
	"strings"

	"github.com/UnivocalX/odessa/internal/storage"
	"github.com/UnivocalX/odessa/internal/storage/azure"
	"github.com/UnivocalX/odessa/internal/storage/fs"
	"github.com/UnivocalX/odessa/internal/storage/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

// ConfigureStorage configures the auto-registered storage backends from config.
func ConfigureStorage(cfg StorageConfig) error {
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

// NewViper creates a pre-configured viper instance with the ODESSA env prefix.
func NewViper() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("ODESSA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	return v
}

// ReadConfigFile reads a YAML config file into the viper instance if specified.
func ReadConfigFile(v *viper.Viper, cfgFile string) error {
	if cfgFile == "" {
		return nil
	}
	v.SetConfigFile(cfgFile)
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read config %q: %w", cfgFile, err)
	}
	return nil
}

// BindStorageFlags binds the common storage CLI flags to viper keys.
func BindStorageFlags(v *viper.Viper, cmd *cobra.Command) {
	bindFlag := func(key, name string) {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			_ = v.BindPFlag(key, flag)
		}
	}
	bindFlag("storage.fs.root", "fs-root")
	bindFlag("storage.s3.region", "s3-region")
	bindFlag("storage.s3.endpoint", "s3-endpoint")
	bindFlag("storage.azure.account", "az-account")
}

// RegisterStorageFlags adds the standard storage backend flags to a cobra command.
func RegisterStorageFlags(cmd *cobra.Command) {
	cmd.Flags().String("fs-root", "", "filesystem backend root directory")
	cmd.Flags().String("s3-region", "", "S3 region (enables S3 backend)")
	cmd.Flags().String("s3-endpoint", "", "S3-compatible endpoint URL")
	cmd.Flags().String("az-account", "", "Azure storage account name (enables Azure backend)")
}
