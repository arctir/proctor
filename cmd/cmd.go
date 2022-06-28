package cmd

import (
	"fmt"
	"os"

	"github.com/arctir/proctor/process"
	"github.com/spf13/cobra"
)

var proctorCmd = &cobra.Command{
	Use:   "proctor",
	Short: "A command-line tool for inspecting processes and understanding their relationships.",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves a process by name.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please enter a process name")
			return
		}
		process.RunGetProcess(args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all available processes known to the system.",
	Run: func(cmd *cobra.Command, args []string) {
		process.RunGetProcesses()
	},
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "List a process and all it's relevant ancestors.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please enter a process name")
			return
		}
		process.RunGetProcessForRelationship(args[0])
	},
}

func SetupCommands() *cobra.Command {
	proctorCmd.AddCommand(listCmd)
	proctorCmd.AddCommand(getCmd)
	proctorCmd.AddCommand(treeCmd)

	if err := proctorCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return proctorCmd
}
