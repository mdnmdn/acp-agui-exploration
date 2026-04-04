package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	agentFlag string
)

var rootCmd = &cobra.Command{
	Use:   "sample-acp",
	Short: "ACP Agent TUI and CLI client",
	Run: func(cmd *cobra.Command, args []string) {
		if agentFlag != "" && len(args) > 0 {
			message := strings.Join(args, " ")
			if err := runCliMode(agentFlag, message); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
		
		// If no CLI args provided, start TUI
		startTui()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&agentFlag, "agent", "a", "", "Agent to use (bypasses TUI if message is provided)")
}

func Execute() {
	if len(os.Args) > 1 && os.Args[1] == "mock-agent" {
		runMockAgent()
		return
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
