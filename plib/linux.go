package plib

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultProcDir = string(os.PathSeparator) + "proc"
	cmdDir         = "cmdline"
	statDir        = "stat"
	exeDir         = "exe"
	nullCharacter  = "\x00"
	permDenied     = "PERM_DENIED"
	statError      = "ERROR_READING_STAT"
)

type LinuxInspector struct{}

// ProcessStat is a representation of procfs's stat file in Linux hosts.
// https://www.kernel.org/doc/html/latest/filesystems/proc.html#id10
type ProcessStat struct {
	// ID of a process (pid)
	ID              int
	FileName        string // tcomm
	State           string // state (R: Running, S: Sleeping, D: Sleeping and uninteruptable, Z: Zombie, T: Traced or stopped)
	ParentID        int    // ppid
	Group           int    // pgrp
	SessionID       int    // sid
	TTY             int    // tty_nr
	TTYProcessGroup int    // tty_pgrp
	// the task's current scheduling flags which are expressed in hexadecimal
	// notation and with zeros suppressed.
	TaskFlags string
	// min_flt
	MinorFaultQuantity          int
	MinorFaultWithChildQuantity int
	MajorFaultQuantity          int
	MajorFaultWithChildQuantity int
	UTime                       int
	KernalTime                  int
	UTimeWithChild              int
	KernalTimeWithChild         int
	Priority                    int
	Nice                        int
	ThreadQuantity              int
	ItRealValue                 int
	// time the process started after boot
	StartTime            int
	VirtualMemSize       int
	ResidentSetMemSize   int
	RSSByteLimit         int
	StartCode            int
	EndCode              int
	StartStack           int
	ESP                  int
	EIP                  int
	PendingSignals       int
	BlockedSignals       int
	IgnoredSignals       int
	CaughtSignals        int
	PlaceHolder1         int
	PlaceHolder2         int
	PlaceHolder3         int
	ExitSignal           int
	CPU                  int
	RealtimePriority     int
	SchedulingPolicy     int
	TimeSpentOnBlockIO   int
	GuestTime            int
	GuestTimeWithChild   int
	StartDataAddress     int
	EndDataAddress       int
	HeapExpansionAddress int
	StartCMDAddress      int
	EndCMDAddress        int
	StartEnvAddress      int
	EndEnvAddress        int
	ExitCode             int
}

type ListOptions struct {
	IncludeKernel           bool
	IncludePermissionIssues bool
}

// TODO(joshrosso)
func (l *LinuxInspector) ListProcesses() ([]Process, error) {
	ps, err := GetProcesses()
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func resolvePIDRelationship(FullPIDList *[]int, pidlist map[int]Process, rootPID int) {
	//fmt.Printf("ID that entered was %d\n", rootPID)
	*FullPIDList = append(*FullPIDList, rootPID)
	parentID := pidlist[rootPID].ParentProcess
	//time.Sleep(1 * time.Second)
	//fmt.Printf("Parent ID was %d\n", parentID)
	if parentID > 1 {
		resolvePIDRelationship(FullPIDList, pidlist, parentID)
	}
	if parentID == 1 {
		*FullPIDList = append(*FullPIDList, 1)
	}
	//fmt.Printf("FINAL: %s\n", FullPIDList)
}

func RunGetProcessForRelationship(name string) ProcessRelation {
	ps, err := GetProcesses()
	if err != nil {
		panic(err)
	}

	processByID := map[int]Process{}
	for i := range ps {
		processByID[ps[i].ID] = ps[i]
	}

	processByName := map[string]Process{}
	for i := range ps {
		if ps[i].CommandName != "" {
			processByName[ps[i].CommandName] = ps[i]
		}
	}

	pidsInvolved := []int{}
	resolvePIDRelationship(&pidsInvolved, processByID, processByName[name].ID)

	processRelations := []ProcessRelation{}
	for _, pid := range pidsInvolved {
		processRelations = append(processRelations, ProcessRelation{Process: processByID[pid]})
	}

	for i := range processRelations {
		if i == len(processRelations)-1 {
			processRelations[i].Process.Stat = nil
			break
		}

		// temp cleaner
		processRelations[i].Process.Stat = nil

		processRelations[i].Parent = &processRelations[i+1]
	}

	return processRelations[0]
}

func addRelativeProcess(ps *ProcessRelation) {
}

// GetProcessesByName looks up a process based on its name. In the case of
// linux, this is done by TODO(joshrosso). An error is returned if process
// lookup failed. If no process with the provided name is found, an empty slice
// is returned.
func GetProcessesByName(name string) ([]Process, error) {
	results := []Process{}
	ps, err := GetProcesses()
	if err != nil {
		return []Process{}, err
	}
	for i := range ps {
		if ps[i].CommandName == name {
			results = append(results, ps[i])
		}
	}

	return results, nil
}

func RunGetProcesses() []Process {
	ps, err := GetProcesses()
	if err != nil {
		//TODO(joshrosso): Deal with this.
		panic(err)
	}

	return ps
}

// getPIDs returns every process ID known to procfs. A process ID is considered
// valid if it is a directory with a numeric name. An error is returned when
// getPIDs is unable to read procfs.
func getPIDs() ([]int, error) {
	procDirs, err := os.ReadDir(defaultProcDir)
	if err != nil {
		return nil, err
	}

	pids := []int{}
	for _, p := range procDirs {
		// When a directory name is not [^0-9], its not a process and is skipped.
		pid, err := strconv.Atoi(p.Name())
		if err != nil {
			//fmt.Printf("WARN: Omitting proc dir entry %s", p.Name())
			break
		}
		pids = append(pids, pid)
	}

	return pids, nil
}

// LoadProcessName returns the name of the process for the provided PID. If the
// name cannot be resolved, an empty string is returned.
func LoadProcessName(pid int) (string, error) {
	path, err := LoadProcessPath(pid)
	if err != nil {
		return "", err
	}
	dirs := strings.Split(path, string(os.PathSeparator))
	if len(dirs) < 1 {
		return "", nil
	}
	return dirs[len(dirs)-1], nil
}

// LoadProcessSHA evaluates the sha256 value of the binary.
func LoadProcessSHA(path string) string {
	if path == "" {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		//TODO(joshrosso): fix this
		panic(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		//TODO(joshrosso): fix this
		panic(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// LoadProcessPath returns the path, or location, of the binary being executed
// as a process. To reliably determine the path, it reads the symbolic link in
// /proc/${PID}/exe and resolves the final file as seperated by "/". While a
// reliable way to approach process resolution on Linux, it does require root
// access to resolve.
//
// TODO(joshrosso): Consider a more logic-based approach to name resolution
// when root access is not possible.
func LoadProcessPath(pid int) (string, error) {
	exeLink, err := os.Readlink(filepath.Join(defaultProcDir, strconv.Itoa(pid), exeDir))
	if err != nil {
		return "", err
	}
	return exeLink, nil
}

// LoadProcessDetails introspects the process's directory in procfs to retrieve
// relevant information and product a instance of Process. The generated
// Process object is then returned. No error is returned, as missing
// information or lack of access to data in procfs will result in missing
// information in the generated returned Process.
func LoadProcessDetails(pid int) Process {
	hasPerm := true
	isK := false
	var sha string
	name, err := LoadProcessName(pid)

	// when error is bubbled up, determine why to set name correctly
	if err != nil {
		switch {
		case os.IsPermission(err):
			name = permDenied
			hasPerm = false
		case os.IsNotExist(err):
			stat, err := os.ReadFile(filepath.Join(defaultProcDir, strconv.Itoa(pid), statDir))
			if err != nil {
				//TODO(joshrosso): Clean this up.
				panic(err)
			}
			//TODO(joshrosso): But does this handle nullCharacter?
			parsedStats := strings.Split(string(stat), " ")
			name = parsedStats[1]
			isK = true
		default:
			name = "ERROR_UNKNOWN"
		}

	}
	path, err := LoadProcessPath(pid)
	//TODO(joshrosso): cleanup this logic.
	if err != nil {
		if os.IsPermission(err) {
			path = permDenied
			sha = permDenied
		} else {
			path = statError
			sha = statError
		}

	} else {
		sha = LoadProcessSHA(path)
	}
	stat := LoadStat(pid)

	p := Process{
		ID:            pid,
		IsKernel:      isK,
		HasPermission: hasPerm,
		CommandName:   name,
		CommandPath:   path,
		ParentProcess: stat.ParentID,
		BinarySHA:     sha,
		Stat:          &stat,
	}

	return p
}

// GetProcesses retrieves all the processes from procfs. It introspects
// each process to gather data and returns a slice of Processes. An error is
// returned when GetProcesses is unable to interact with procfs.
func GetProcesses(opts ...ListOptions) ([]Process, error) {
	opt := MergeOptions(opts)
	pids, err := getPIDs()
	if err != nil {
		return nil, err
	}

	procs := []Process{}

	for _, pid := range pids {
		p := LoadProcessDetails(pid)
		switch {
		// filter out kernel processes and permission issues
		case !opt.IncludeKernel && !opt.IncludePermissionIssues:
			if !p.IsKernel && p.HasPermission {
				procs = append(procs, p)
			}
		// filter out permission issues, include kernel processes
		case opt.IncludeKernel && !opt.IncludePermissionIssues:
			if p.HasPermission {
				procs = append(procs, p)
			}
		// filter out kernel processes, include permission issues
		case !opt.IncludeKernel && opt.IncludePermissionIssues:
			if !p.IsKernel {
				procs = append(procs, p)
			}
		// include all processes
		case opt.IncludeKernel && opt.IncludePermissionIssues:
			procs = append(procs, p)
		}
	}

	return procs, nil
}

func MergeOptions(opts []ListOptions) ListOptions {
	// default case when opts are empty
	if len(opts) < 1 {
		return ListOptions{}
	}
	// TODO(joshrosso): Need to do actual merge logic rather than perferring
	// first option
	return opts[0]
}

// LoadStat translates fields in the stat file (/proc/${PID}/stat) into
// structured data. Details on stat contents can be found at
// https://www.kernel.org/doc/html/latest/filesystems/proc.html#id10.
func LoadStat(pid int) ProcessStat {
	ps := ProcessStat{}
	stat, err := os.ReadFile(filepath.Join(defaultProcDir, strconv.Itoa(pid), statDir))
	if err != nil {
		return ps
	}
	parsedStats := strings.Split(string(stat), " ")
	ps.ParentID, _ = strconv.Atoi(parsedStats[3])

	for i, stat := range parsedStats {
		switch i {
		case 0:
			ps.ID, _ = strconv.Atoi(stat)
		case 1:
			ps.FileName = stat
		}

	}

	return ps
}
