package main

import (
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	slog.SetDefault(logger)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
