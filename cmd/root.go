package cmd

import (
	"fmt"
	"os"

	"github.com/pprunty/magikarp/pkg/terminal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "magikarp",
	Short: "Magikarp - AI Coding Assistant CLI",
	Long: `Magikarp is an open-source coding assistant CLI tool built with Go. 
It provides an interactive terminal interface for AI-powered coding assistance 
with support for multiple LLM providers including Claude, GPT, and Gemini.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check terminal capabilities before starting UI
		if err := terminal.CheckTerminalCapabilities(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Start the interactive UI
		if err := terminal.StartUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error starting UI: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.magikarp.yaml)")
}