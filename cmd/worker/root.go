package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnivocalX/odessa/internal/config"
	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/storage"
	"github.com/UnivocalX/odessa/internal/tasks"

	_ "github.com/UnivocalX/odessa/internal/storage/azure"
	_ "github.com/UnivocalX/odessa/internal/storage/fs"
	_ "github.com/UnivocalX/odessa/internal/storage/s3"

	"github.com/spf13/cobra"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "worker",
	Short: "Odessa background worker",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd, configFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	repo, err := repository.New(repository.Config{DSN: cfg.DSN})
	if err != nil {
		slog.Error("open database", "error", err)
		return err
	}
	defer repo.Close()

	if err := config.ConfigureStorage(cfg.Storage); err != nil {
		return fmt.Errorf("configure storage: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := NewWorker(WorkerConfig{
		PollInterval:    cfg.PollInterval,
		MaxPollInterval: cfg.MaxPollInterval,
		Concurrency:     cfg.Concurrency,
		MaxAttempts:     cfg.MaxAttempts,
		DrainTimeout:    cfg.DrainTimeout,
	})

	// Register task handlers.
	w.Register(tasks.NewScanOriginHandler(repo, storage.Default()))

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		slog.Info("shutting down worker, draining in-flight tasks...", "drain_timeout", cfg.DrainTimeout)
		cancel()
	}()

	slog.Info("worker started",
		"poll_interval", cfg.PollInterval,
		"max_poll_interval", cfg.MaxPollInterval,
		"concurrency", cfg.Concurrency,
		"max_attempts", cfg.MaxAttempts,
		"drain_timeout", cfg.DrainTimeout,
	)
	w.Run(ctx)
	slog.Info("worker stopped")
	return nil
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "path to YAML config file")
	rootCmd.Flags().String("dsn", "", "PostgreSQL connection string (required)")
	rootCmd.Flags().Duration("poll-interval", 5*time.Second, "base interval between polling for new tasks")
	rootCmd.Flags().Duration("max-poll-interval", 2*time.Minute, "maximum backoff interval when idle")
	rootCmd.Flags().Int("concurrency", 4, "number of concurrent task processors")
	rootCmd.Flags().Int("max-attempts", 3, "maximum retry attempts before marking a scan as failed")
	rootCmd.Flags().Duration("drain-timeout", 30*time.Second, "max time to wait for in-flight tasks on shutdown")

	config.RegisterStorageFlags(rootCmd)
}
