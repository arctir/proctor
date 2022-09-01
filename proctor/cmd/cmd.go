package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/arctir/proctor/plib"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	// output
	getCmd.Flags().StringP(outputFlag, "o", "table", "Output type for command [table (default), json].")
	listCmd.Flags().StringP(outputFlag, "o", "table", "Output type for command [table (default), json].")
	treeCmd.Flags().StringP(outputFlag, "o", "table", "Output type for command [table (default), json].")

	// kernel filter
	getCmd.Flags().Bool(includeKernelFlag, false, "Include kernel processes in out, default is false.")
	listCmd.Flags().Bool(includeKernelFlag, false, "Include kernel processes in out, default is false.")
	treeCmd.Flags().Bool(includeKernelFlag, false, "Include kernel processes in out, default is false.")

	// permission filter
	listCmd.Flags().Bool(includePermIssueFlag, false, "Include processes that proctor failed to introspect due to permission issues.")
	treeCmd.Flags().Bool(includePermIssueFlag, false, "Include processes that proctor failed to introspect due to permission issues.")
	getCmd.Flags().Bool(includePermIssueFlag, false, "Include processes that proctor failed to introspect due to permission issues.")
}

type OutputType int

const (
	jsonOut OutputType = iota
	tableOut
	outputFlag           = "output"
	includeKernelFlag    = "include-kernel"
	includePermIssueFlag = "include-permission-issues"
)

type ProctorOpts struct {
	outType          OutputType
	includeKernel    bool
	includePermIssue bool
}

var proctorCmd = &cobra.Command{
	Use:   "proctor",
	Short: "A command-line tool for inspecting processes and understanding their relationships.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves a process's details.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please enter a process name")
			return
		}
		AssembleGetProcesses(args[0], newOptions(cmd.Flags()))
	},
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Retrieve a process's and it's ancestor(s)' details.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please enter a process name")
			return
		}
		AssembleTreeForProcess(args[0], newOptions(cmd.Flags()))
	},
}

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all available processes and their details.",
	Run: func(cmd *cobra.Command, args []string) {
		AssembleListProcesses(newOptions(cmd.Flags()))
	},
}

// SetupCommands adds the CLI commands to the proctor CLI.
func SetupCommands() *cobra.Command {
	proctorCmd.AddCommand(listCmd)
	proctorCmd.AddCommand(getCmd)
	proctorCmd.AddCommand(treeCmd)

	return proctorCmd
}

func AssembleGetProcesses(processName string, opts ProctorOpts) {
	plibOpts := plib.ListOptions{
		IncludeKernel:           opts.includeKernel,
		IncludePermissionIssues: opts.includePermIssue,
	}

	var out []byte
	ps, err := plib.GetProcessesByName(processName, plibOpts)
	// TODO(joshrosso): deal with panic
	if err != nil {
		panic(err.Error())
	}

	switch opts.outType {
	case jsonOut:
		out, err = json.Marshal(ps)
		// TODO(joshrosso): Make this better.
		if err != nil {
			panic(err.Error())
		}
	default:
		out = createGetTable(ps)
	}

	fmt.Printf("%s", out)
}

// AssembleTreeForProcess derives a tree of ancestor processes by introspecting
// TODO(joshrosso): move away from outputOption type string
func AssembleTreeForProcess(processName string, opts ProctorOpts) {
	plibOpts := plib.ListOptions{
		IncludeKernel:           opts.includeKernel,
		IncludePermissionIssues: opts.includePermIssue,
	}

	psRelationships := plib.RunGetProcessForRelationship(processName, plibOpts)
	var out []byte
	var err error
	switch opts.outType {
	case jsonOut:
		out, err = json.Marshal(psRelationships)
		// TODO(joshrosso): Make this better.
		if err != nil {
			panic(err.Error())
		}
	default:
		out = createTreeTable(psRelationships)
	}
	fmt.Printf("%s", out)
}

func AssembleListProcesses(opts ProctorOpts) {
	plibOpts := plib.ListOptions{
		IncludeKernel:           opts.includeKernel,
		IncludePermissionIssues: opts.includePermIssue,
	}

	ps, err := plib.GetProcesses(plibOpts)
	if err != nil {
		// TODO(joshrosso): Make this better.
		panic(err.Error())
	}

	var out []byte
	switch opts.outType {
	case jsonOut:
		out, err = json.Marshal(ps)
		// TODO(joshrosso): Make this better.
		if err != nil {
			panic(err.Error())
		}
	default:
		out = createGetTable(ps)
		fmt.Printf("%s", out)
	}
}

func createTreeTable(pr plib.ProcessRelation) []byte {
	ps := [][]string{}
	prte := pr
	for {
		ps = append(ps, []string{strconv.Itoa(prte.Process.ID), prte.Process.CommandName, prte.Process.CommandPath, prte.Process.BinarySHA})
		if prte.Parent == nil {
			break
		}
		prte = *prte.Parent
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"PID", "name", "location", "SHA"})
	table.AppendBulk(ps)
	table.Render()
	return buf.Bytes()
}

func createGetTable(ps []plib.Process) []byte {
	psList := [][]string{}
	for _, p := range ps {
		psList = append(psList, []string{strconv.Itoa(p.ID), p.CommandName, p.CommandPath, p.BinarySHA})
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"PID", "name", "location", "SHA"})
	table.AppendBulk(psList)
	table.Render()
	return buf.Bytes()
}

func newOptions(fs *pflag.FlagSet) ProctorOpts {
	ot := resolveOutputType(fs)
	fko, err := fs.GetBool(includeKernelFlag)
	if err != nil {
		fko = false
	}
	ipi, err := fs.GetBool(includePermIssueFlag)
	if err != nil {
		ipi = false
	}

	return ProctorOpts{
		outType:          ot,
		includeKernel:    fko,
		includePermIssue: ipi,
	}
}

func resolveOutputType(fs *pflag.FlagSet) OutputType {
	of, err := fs.GetString(outputFlag)
	// default if there are ever issues finding flag
	if err != nil {
		return tableOut
	}
	switch of {
	case "json":
		return jsonOut
	case "table":
		return tableOut
	}

	// default OutputType
	return tableOut
}
