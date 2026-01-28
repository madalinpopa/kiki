package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	logDirName  = ".kiki"
	logFileName = "kiki.log"
	logDirPerm  = 0o755
	logFilePerm = 0o644
)

// GetLogDir returns the directory used for log files (~/.kiki).
func GetLogDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(homeDir, logDirName), nil
}

// NewFileLogger creates a slog logger that writes to ~/.kiki/kiki.log.
func NewFileLogger() (*slog.Logger, io.Closer, error) {
	logDir, err := GetLogDir()
	if err != nil {
		return nil, nil, err
	}

	if err := os.MkdirAll(logDir, logDirPerm); err != nil {
		return nil, nil, fmt.Errorf("create log dir: %w", err)
	}

	logPath := filepath.Join(logDir, logFileName)
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePerm)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	handler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return slog.New(handler), file, nil
}
