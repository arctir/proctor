// plib is Arctir's process library. This library is used to gather process
// details across a variety of operating systems.
package plib

import (
	"fmt"
	"runtime"
)

type Inspector interface {
	ListProcesses() ([]Process, error)
}

// ProcessQueryOptions are used when looking up a process. The fields act as
// filters that can be applied.
type ProcessQueryOptions struct {
	ProcessName string
	ProcessID   int
}

type ProcessRelation struct {
	Process Process
	Parent  *ProcessRelation
}

type Process struct {
	ID            int
	BinarySHA     string
	CommandName   string
	CommandPath   string
	FlagsAndArgs  string
	ParentProcess int
	IsKernel      bool
	HasPermission bool
	Stat          any
}

// NewInspector returns a an Inspector instance based on the host's operating system. If the host's operating system cannot be detected or the operating system is unsupported, an error is returned.
func NewInspector() (Inspector, error) {
	switch runtime.GOOS {
	// TODO(joshrosso): Other target architectures
	case "linux":
		return &LinuxInspector{}, nil
	}

	return nil, fmt.Errorf("failed to create inspector because operating system %s is unsupported\n", runtime.GOOS)
}
