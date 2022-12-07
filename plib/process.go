// plib is a library for efficiently retrieving information about processes on
// various operating systems.
package plib

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/adrg/xdg"
)

// plib defaults
const (
	// The name directory containing the cache file. Note the base
	CacheDirName = "proctor"
	// The name of the file storing a cache of loaded process details.
	CacheFileName = "proc.cache"
)

// Process is an operating system's representation of execution, typically
// referred to as a process. Details available about processes vary between
// operating systems. As such, Process contains multiple common fields that are
// resolvable in most operating systems and a field for operating
// system-specific details. While this value can be any type, the Type field in
// Process can be used to determine which operating system the process
// origniated from, which can then be used to inform casting of the interface
// (any) to a concrete type.
type Process struct {
	// The process's numeric identifier. On *nix, this is the pid.
	ID int
	// The SHA256 value representing the process's binary.
	BinarySHA string
	// The name of the command that was triggered for execution.
	CommandName string
	// The full path of the command (binary) that was run.
	CommandPath  string
	FlagsAndArgs string
	// The parent process's numeric identifier. The parent is typically the
	// process which kicked off this process.
	ParentProcess int
	// Whether this is a standard process or kernel-level process. For example,
	// in Linux, a standard process would be considered one that is managed in
	// userspace relative to being managed by the Linux kernel.
	IsKernel bool
	// Whether the user used to access this process's details has permissions to
	// gather them.
	HasPermission bool
	// Type is the operating-system type this process was retrieved on. This
	// value can be used to cast the OSSpecfic field below so that you can access
	// its fields. For example, p.OSSpecific.(LinuxStat).
	Type string
	// Additional process details that are specific to the underlying operating
	// system. The fields found here will vary depending on whether processes
	// were retrieved on Linux, Windows, BSD, etc. For client's accessing this
	// field, use the above Type field to determine which operating system the
	// process orginigated from to inform casting this interface to a concrete
	// struct type.
	OSSpecific any
}

// Processes is a map of Process pointers where the key is the ID of each
// process. This facilitates easier lookup and relation mapping (e.g.
// determining a process's parent) for the caller.
type Processes map[int]*Process

// Inspector is able to load, cache, and return all the process information
// available to the user on an operating system. Each implementation represents
// a unique operating system.
type Inspector interface {
	// LoadProcesses gathers all the available process information from the host
	// using its available API(s). For example, on a Linux host, LoadProcesses
	// may load process details from the process virtual filesystem (procfs). If
	// the implementator supports caching, and the client wishes to cache,
	// LoadProcesses may also persist the process information for future
	// operations. An error is retruned if process information is unable to be
	// looked up using the operating system's APIs.
	LoadProcesses() error
	// ClearProcessCache will empty any cached process information stored from a
	// LoadProcesses call. If called when there is no cache to clear, it returns
	// without error. An error is returned if a cache is available to clear but
	// it is unable to do so.
	ClearProcessCache() error
	// GetProcesses retrieves all process information available. If a cache
	// pre-exists, GetProcesses will load the processes from that cache. If a
	// cache does not exist, the implementation should run LoadProcesses and
	// return the result. Whether a new cache is formed in this situation is up
	// to the implementator and, potentially, its configuration.
	GetProcesses() (Processes, error)
	// NOTE(joshrosso): I'm considering expansion of this interface over time.
	// While it's compelling to include "helper" functions like GetProcessByName
	// it also could cause an expansion of endless possibilities, when the point
	// of the inspector should be to efficienlty resolve, cache, and return
	// processes. It is, IMO, easy for the client to make alternative structures
	// from the returned processes that can facilitate its exact lookup needs.
}

type InspectorConfig struct {
	// The location on the filesystem where process details can be cached. By
	// default this will be set to $XDG_DATA_HOME/proctor/proc.cache. Unless you
	// are writing tests, you probably should not set this value.
	CacheFilePath string
	// Whether an existing cache should be ignored thus retrieving all process's
	// from the operating system's API(s).
	IgnoreCache bool
	// OSSpecificConfig enables you to provide specific settings that are only
	// relevant to a specific Inspector implementation. For example, if you
	// wanted to provide Linux-specific provide configuration, you could setup a
	// [LinuxInspectorConfig] and place it here.
	OSSpecificConfig any
}

// NewInspector returns an Inspector instance based on the host's operating
// system. If the host's operating system cannot be detected or the operating
// system is unsupported, an error is returned.
func NewInspector() (Inspector, error) {
	switch runtime.GOOS {
	// TODO(joshrosso): Other target architectures
	case "linux":
		return nil, nil
		//return &LinuxInspector{}, nil
	}

	return nil, fmt.Errorf(
		"failed to create inspector because operating system %s is unsupported\n",
		runtime.GOOS,
	)
}

// GetCacheLocation returns the location where process details can be cached.
// This will resolve to the caller's equivalent of
// $XDG_DATA_HOME/CacheDirName/CacheFileName.
func GetCacheLocation() string {
	return filepath.Join(xdg.DataHome, CacheDirName, CacheFileName)
}
