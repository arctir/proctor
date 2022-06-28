package process

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ProcessRelation struct {
	Process Process
	Parent  *ProcessRelation
}

type Process struct {
	ID            int
	CommandName   string
	CommandPath   string
	FlagsAndArgs  string
	ParentProcess int
	Stat          *ProcessStat
}

// ProcessStat is a representation of procfs's stat file.
// https://www.kernel.org/doc/html/latest/filesystems/proc.html#id10
type ProcessStat struct {
	ID              int    // pid
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

const (
	procDir       = string(os.PathSeparator) + "proc"
	cmdDir        = "cmdline"
	statDir       = "stat"
	exeDir        = "exe"
	nullCharacter = "\x00"
)

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

func RunGetProcessForRelationship(name string) {
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

	for i, _ := range processRelations {
		if i == len(processRelations)-1 {
			processRelations[i].Process.Stat = nil
			break
		}

		// temp cleaner
		processRelations[i].Process.Stat = nil

		processRelations[i].Parent = &processRelations[i+1]
	}

	d, _ := json.Marshal(processRelations[0])
	fmt.Println(string(d))
	//fmt.Println(pidsInvolved)
}

func addRelativeProcess(ps *ProcessRelation) {
}

func RunGetProcess(name string) {
	ps, err := GetProcesses()
	if err != nil {
		panic(err)
	}
	processByName := map[string]Process{}
	for i := range ps {
		if ps[i].CommandName != "" {
			processByName[ps[i].CommandName] = ps[i]
		}
	}

	if val, ok := processByName[name]; ok {
		d, _ := json.Marshal(val)
		fmt.Println(string(d))
	} else {
		fmt.Printf("No process named %s found.", name)
	}

}

func RunGetProcesses() {
	ps, err := GetProcesses()
	if err != nil {
		panic(err)
	}

	d, _ := json.Marshal(ps)
	fmt.Println(string(d))
}

// getPIDs returns every process ID known to procfs. A process ID is considered
// valid if it is a directory with a numeric name. An error is returned when
// getPIDs is unable to read procfs.
func getPIDs() ([]int, error) {
	procDirs, err := os.ReadDir(procDir)
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
func LoadProcessName(pid int) string {
	path := LoadProcessPath(pid)
	dirs := strings.Split(path, string(os.PathSeparator))
	if len(dirs) < 1 {
		return ""
	}
	return dirs[len(dirs)-1]
}

// LoadProcessPath returns the path, or location, of the binary being executed
// as a process. To reliably determine the path, it reads the symbolic link in
// /proc/${PID}/exe and resolves the final file as seperated by "/". While a
// reliable way to approach process resolution on Linux, it does require root
// access to resolve.
//
// TODO(joshrosso): Consider a more logic-based approach to name resolution
// when root access is not possible.
func LoadProcessPath(pid int) string {
	exeLink, err := os.Readlink(filepath.Join(procDir, strconv.Itoa(pid), exeDir))
	if err != nil {
		//fmt.Printf("WARN: Could not read the link at /proc/%d/exe\n", pid)
	}
	return exeLink
}

// LoadProcessDetails introspects the process's directory in procfs to retrieve
// relevant information and product a instance of Process. The generated
// Process object is then returned. No error is returned, as missing
// information or lack of access to data in procfs will result in missing
// information in the generated returned Process.
func LoadProcessDetails(pid int) Process {
	name := LoadProcessName(pid)
	path := LoadProcessPath(pid)
	stat := LoadStat(pid)

	p := Process{
		ID:            pid,
		CommandName:   name,
		CommandPath:   path,
		ParentProcess: stat.ParentID,
		Stat:          &stat,
	}

	return p
}

// GetProcesses retrieves all the processes from procfs. It introspects
// each process to gather data and returns a slice of Processes. An error is
// returned when GetProcesses is unable to interact with procfs.
func GetProcesses() ([]Process, error) {
	pids, err := getPIDs()
	if err != nil {
		return nil, err
	}

	procs := []Process{}
	for _, pid := range pids {
		procs = append(procs, LoadProcessDetails(pid))
	}

	return procs, nil
}

// LoadStat translates fields in the stat file (/proc/${PID}/stat) into
// structured data. Details on stat contents can be found at
// https://www.kernel.org/doc/html/latest/filesystems/proc.html#id10.
func LoadStat(pid int) ProcessStat {
	ps := ProcessStat{}
	stat, err := os.ReadFile(filepath.Join(procDir, strconv.Itoa(pid), statDir))
	if err != nil {
		return ps
	}
	parsedStats := strings.Split(string(stat), " ")
	//	fmt.Printf("LOOKING UP PID %d\n", pid)
	//for i, statVal := range parsedStats {
	//		fmt.Printf("%d: %s\n", i, statVal)
	//}
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
