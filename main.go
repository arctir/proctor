package main

import (
	"fmt"
	"os"

	"github.com/arctir/proctor/cmd"
)

func main() {
	proctorCmd := cmd.SetupCommands()
	if err := proctorCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
