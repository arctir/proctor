package main

import (
	"fmt"
	"os"

	"github.com/arctir/proctor/proctor/cmd"
)

func main() {
	proctorCmd := cmd.SetupCLI()
	if err := proctorCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
