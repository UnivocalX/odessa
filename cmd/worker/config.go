package main

import (
	"fmt"
	"time"

	"github.com/UnivocalX/odessa/internal/core"
	"github.com/spf13/cobra"
)

type Config struct {
	DSN             core.Secret        `mapstructure:"dsn" validate:"required"`
	PollInterval    time.Duration      `mapstructure:"poll_interval"`
	MaxPollInterval time.Duration      `mapstructure:"max_poll_interval"`
	Concurrency     int                `mapstructure:"concurrency" validate:"gt=0"`
	MaxAttempts     int                `mapstructure:"max_attempts" validate:"gt=0"`
	DrainTimeout    time.Duration      `mapstructure:"drain_timeout"`
	Storage         core.StorageConfig `mapstructure:"storage"`
}

func loadConfig(cmd *cobra.Command, cfgFile string) (Config, error) {
	v := core.NewViper()

	v.SetDefault("dsn", "")
	v.SetDefault("poll_interval", 5*time.Second)
	v.SetDefault("max_poll_interval", 2*time.Minute)
	v.SetDefault("concurrency", 4)
	v.SetDefault("max_attempts", 3)
	v.SetDefault("drain_timeout", 30*time.Second)

	if err := core.ReadConfigFile(v, cfgFile); err != nil {
		return Config{}, err
	}

	bindFlag := func(key, name string) {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			_ = v.BindPFlag(key, flag)
		}
	}
	bindFlag("dsn", "dsn")
	bindFlag("poll_interval", "poll-interval")
	bindFlag("max_poll_interval", "max-poll-interval")
	bindFlag("concurrency", "concurrency")
	bindFlag("max_attempts", "max-attempts")
	bindFlag("drain_timeout", "drain-timeout")

	core.BindStorageFlags(v, cmd)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}
