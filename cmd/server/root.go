package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/UnivocalX/odessa/internal/repository"
	"github.com/UnivocalX/odessa/internal/server"
	"github.com/UnivocalX/odessa/internal/service"
	"github.com/UnivocalX/odessa/internal/storage"

	_ "github.com/UnivocalX/odessa/internal/storage/azure"
	_ "github.com/UnivocalX/odessa/internal/storage/fs"
	_ "github.com/UnivocalX/odessa/internal/storage/s3"

	"github.com/spf13/cobra"
)

var configFile string

func newEmailSender(cfg SMTPConfig) service.EmailSender {
	if cfg.Host == "" || cfg.From == "" {
		return nil
	}
	return service.NewSMTPEmailSender(service.SMTPConfig{
		Host: cfg.Host, Port: cfg.Port, Username: cfg.Username,
		Password: cfg.Password.Expose(), From: cfg.From,
	})
}

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Odessa HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd, configFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		repo, err := repository.New(repository.Config{DSN: cfg.DSN})
		if err != nil {
			slog.Error("open database", "error", err)
			return err
		}

		svc := service.New(repo, storage.Default(), service.AuthOptions{
			JWTSecret:            cfg.Auth.JWTSecret,
			AccessTokenLifetime:  cfg.Auth.AccessTokenLifetime,
			RefreshTokenLifetime: cfg.Auth.RefreshTokenLifetime,
			ResetTokenLifetime:   cfg.Auth.ResetTokenLifetime,
			PasswordResetURL:     cfg.Auth.PasswordResetURL,
			EmailSender:          newEmailSender(cfg.Email.SMTP),
		})
		created, password, err := svc.EnsureDefaultAdmin(context.Background())
		if err != nil {
			return fmt.Errorf("bootstrap administrator: %w", err)
		}
		if created {
			slog.Warn("default administrator created; save these credentials", "email", service.DefaultAdminEmail, "password", password)
		}

		if err := configureStorage(cfg.Storage); err != nil {
			return fmt.Errorf("configure storage: %w", err)
		}

		srv := server.New(repo, storage.Default(), server.Config{
			Addr: cfg.Addr,
			HTTP: server.Options{
				ReadTimeout:         cfg.HTTP.ReadTimeout,
				WriteTimeout:        cfg.HTTP.WriteTimeout,
				IdleTimeout:         cfg.HTTP.IdleTimeout,
				MaxHeaderBytes:      cfg.HTTP.MaxHeaderBytes,
				MaxRequestBodyBytes: cfg.HTTP.MaxRequestBodyBytes,
			},
			Auth: service.AuthOptions{
				JWTSecret:            cfg.Auth.JWTSecret,
				AccessTokenLifetime:  cfg.Auth.AccessTokenLifetime,
				RefreshTokenLifetime: cfg.Auth.RefreshTokenLifetime,
				ResetTokenLifetime:   cfg.Auth.ResetTokenLifetime,
				PasswordResetURL:     cfg.Auth.PasswordResetURL,
				EmailSender:          newEmailSender(cfg.Email.SMTP),
			},
		})

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("server failed", "error", err)
				os.Exit(1)
			}
		}()

		slog.Info("server started", "addr", cfg.Addr)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		slog.Info("shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown failed", "error", err)
			os.Exit(1)
		}

		slog.Info("server stopped")
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "path to YAML config file")
	rootCmd.Flags().StringP("addr", "a", ":8080", "address to listen on (e.g. :8080 or 0.0.0.0:9090)")
	rootCmd.Flags().String("dsn", "", "PostgreSQL connection string (required)")

	// Storage backend flags — each enables/overrides the corresponding backend.
	// These are merged over any values from the YAML config file.
	rootCmd.Flags().String("fs-root", "", "filesystem backend root directory")
	rootCmd.Flags().String("s3-region", "", "S3 region (enables S3 backend)")
	rootCmd.Flags().String("s3-endpoint", "", "S3-compatible endpoint URL (optional, enables S3 backend)")
	rootCmd.Flags().String("az-account", "", "Azure storage account name (enables Azure backend)")
}
