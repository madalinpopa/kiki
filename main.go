package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	prompt  string
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
		return runPrompt(prompt)
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

func init() {
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Send a prompt to Kiki")
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(exitFailureCode)
	}
}

func runPrompt(prompt string) error {
	storage, err := NewStorage()
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}

	kiki, err := NewKiki(storage)
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

	fmt.Printf("✅ Kiki initialized successfully!\n")
	fmt.Printf("📁 Config directory: %s\n", configDir)
	fmt.Printf("📝 Tasks file: %s/tasks.json\n", configDir)
	fmt.Printf("📝 Notes file: %s/notes.json\n", configDir)
	return nil
}
