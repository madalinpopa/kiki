package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version   = "dev"
	prompt    string
	appLogger *slog.Logger
)

const exitFailureCode = 1

var rootCmd = &cobra.Command{
	Use:   "kiki",
	Short: "Kiki - Your sarcastic personal assistant",
	Long: `Kiki is a sarcastic but helpful CLI assistant for managing tasks and notes.
	
Examples:
  kiki -p "add task: buy milk tomorrow"
  kiki -p "list my tasks"
  kiki -p "what did I note about the API?"
  kiki init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if prompt == "" {
			return cmd.Help()
		}
		return runPrompt(appLogger, prompt)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Kiki configuration",
	Long:  `Creates the Kiki configuration directory and initializes required files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Kiki version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Start a new Kiki session",
	Long:  "Forces Kiki to start a fresh Copilot session for today.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRefresh()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Send a prompt to Kiki")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(refreshCmd)
}

func main() {
	logger, closer, err := NewFileLogger()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(exitFailureCode)
	}
	defer func() {
		if err := closer.Close(); err != nil {
			logger.Error("failed to close log file", "error", err)
		}
	}()

	appLogger = logger
	slog.SetDefault(logger)

	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic", "panic", r, "stack", string(debug.Stack()))
			if _, err := fmt.Fprintln(os.Stderr, "unexpected error"); err != nil {
				logger.Error("failed to write panic message", "error", err)
			}
			os.Exit(exitFailureCode)
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		logger.Error("command failed", "error", err)
		if _, printErr := fmt.Fprintln(os.Stderr, err); printErr != nil {
			logger.Error("failed to write error", "error", printErr)
		}
		os.Exit(exitFailureCode)
	}
}

func runPrompt(logger *slog.Logger, prompt string) error {
	storage, err := NewStorage(logger)
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	kiki, err := NewKiki(storage, logger)
	if err != nil {
		return fmt.Errorf("initializing Kiki: %w", err)
	}
	defer kiki.Close()

	_, err = kiki.Run(prompt)
	if err != nil {
		return fmt.Errorf("running prompt: %w", err)
	}
	return nil
}

func runInit() error {
	configDir := GetConfigDir()

	if err := InitStorage(); err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	if _, err := fmt.Fprintln(os.Stdout, "✅ Kiki initialized successfully!"); err != nil {
		return fmt.Errorf("writing init output: %w", err)
	}
	if _, err := fmt.Fprintf(os.Stdout, "📁 Config directory: %s\n", configDir); err != nil {
		return fmt.Errorf("writing init output: %w", err)
	}
	if _, err := fmt.Fprintf(os.Stdout, "📝 Tasks file: %s/tasks.json\n", configDir); err != nil {
		return fmt.Errorf("writing init output: %w", err)
	}
	if _, err := fmt.Fprintf(os.Stdout, "📝 Notes file: %s/notes.json\n", configDir); err != nil {
		return fmt.Errorf("writing init output: %w", err)
	}
	return nil
}

func runRefresh() error {
	storage, err := NewStorage(appLogger)
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	kiki, err := NewKiki(storage, appLogger)
	if err != nil {
		return fmt.Errorf("initializing Kiki: %w", err)
	}
	defer kiki.Close()

	found, err := kiki.RefreshSession()
	if err != nil {
		return fmt.Errorf("refreshing session: %w", err)
	}

	message := "Kiki session refreshed."
	if !found {
		message = "No existing session found. Next prompt will start a fresh session."
	}
	if _, err := fmt.Fprintln(os.Stdout, message); err != nil {
		return fmt.Errorf("writing refresh output: %w", err)
	}
	return nil
}
