package cmd

import (
	"github.com/spf13/cobra"
)

var proctorCmd = &cobra.Command{
	Use:   "proctor",
	Short: "A command-line tool for inspecting software, from source to runtime.",
	Run:   runProctor,
}

var processCmd = &cobra.Command{
	Use:     "process",
	Aliases: []string{"ps"},
	Short:   "Instrospect processes and understand their relationships.",
	Run:     runProcess,
}

var sourceCmd = &cobra.Command{
	Use:     "source",
	Aliases: []string{"src"},
	Short:   "Introspect source repositories.",
	Run:     runSource,
}

var commitCmd = &cobra.Command{
	Use:   "commits",
	Short: "Access commit details within a repository.",
	Run:   runContrib,
}

var artifactsCmd = &cobra.Command{
	Use:     "artifacts",
	Aliases: []string{"art"},
	Short:   "Artifacts associated with the repository.",
	Run:     runArtifacts,
}

var artifactsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lists all artifacts in a given repository",
	Run:     runListArtifacts,
}

var artifactsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Gets all artifacts for a tag, using the --tag flag.",
	Run:   runGetArtifacts,
}

var contribListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all contributions that have occured in this repository.",
	Run:     runContribList,
}

var contribDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Retrieve the contribution differences between two tags.",
	Run:   runDiffSource,
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available processes and their details.",
	Run:     runListProcesses,
}

var getCmd = &cobra.Command{
	Use:   "get [--name or --id flag]",
	Short: "Retrieves a process's details.",
	Run:   runGetProcess,
}

var treeCmd = &cobra.Command{
	Use:   "tree [pid]",
	Short: "Retrieve a process and all its relatives. Takes a process ID.",
	Run:   runTreeProcess,
}

var fpCmd = &cobra.Command{
	Use:     "finger-print",
	Aliases: []string{"fp"},
	Short:   "Provides a unique checksum representing the process's binary and its parents' binaries combined.",
	Run:     runFingerPrintProcess,
}
