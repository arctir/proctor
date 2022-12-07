package plib

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NewLinuxInspector takes an optional [LinuxInspectorConfig] and returns a
// configured LinuxInspector, which can be used to operator on processes with
// functions like [LinuxInspector.GetProcesses].
//
// The variadic nature of opts is only present to
// make this argument optional. Do not pass multiple opts arguments to this
// function. If you do, the last opt argument passed will be used.
//
// For any required confiuration not specified in the opts argument, including
// if opts is nil, defaults will be set.
func NewLinuxInspector(opts ...LinuxInspectorConfig) LinuxInspector {
	var config LinuxInspectorConfig
	// if opts was passed, used the last indexed argument
	if len(opts) > 0 {
		config = opts[len(opts)-1]
	}

	// TODO(joshrosso): Validate the config details

	return LinuxInspector{
		LinuxInspectorConfig: config,
	}
}

func (l *LinuxInspector) LoadProcesses() error {
	// Also reset the reference to in-memory cache of processes
	l.ps = Processes{}
	ps, err := getPIDsFromProcfs(l.ProcfsFilePath)
	if err != nil {
		return err
	}

	// for each pid, load its data
	for _, p := range ps {
		loadedProcess := LoadProcessStat(l.ProcfsFilePath, p)
		l.ps[p] = &loadedProcess
	}

	// if config says to ignore cache, then exit here.
	if l.IgnoreCache {
		return nil
	}

	err = encodeProcessCache(l.CacheFilePath, l.ps)
	if err != nil {
		return fmt.Errorf("failed persisting process details (cache) to filesystem: %s", err)
	}
	return nil
}

func (l *LinuxInspector) ClearProcessCache() error {
	err := clearProcessCache(l.CacheFilePath)
	if err != nil {
		return fmt.Errorf("failed to clear the existing process cache: %s", err)
	}
	return nil
}

func (l *LinuxInspector) GetProcesses() (Processes, error) {
	// if processes aren't already attached to LinuxInspector, attempt to load
	// them from filesystem cache.
	if l.ps == nil {
		if !l.IgnoreCache {
			l.ps = loadProcessesFromCache(l.CacheFilePath)
		}
	}

	// if processes still aren't loaded into LinuxInspector, attempt to load them
	// from the operating system's API(s).
	if l.ps == nil {
		err := l.LoadProcesses()
		if err != nil {
			return nil, fmt.Errorf("error occured during process retrieval: %s", err)
		}
	}

	if l.ps == nil {
		return nil,
			fmt.Errorf("error occured after process retrieval, where processes still came back nil. This is an unexpected error that should not have occured.")
	}

	return l.ps, nil
}

// loadProcessesFromCache attempts to load processes from an existing cache on
// the filesystem. If there is any form of failure in retrieving processes from
// the cache file, nil is returned.
func loadProcessesFromCache(cacheFp string) Processes {
	var ps Processes

	// need to register ProcessStat as it'll come in as an interface within
	// Process.
	gob.Register(ProcessStat{})

	cacheFileFp := filepath.Join(cacheFp, CacheFileName)
	cacheFile, err := os.Open(cacheFileFp)
	if err != nil {
		return nil
	}
	defer cacheFile.Close()

	encoder := gob.NewDecoder(cacheFile)
	err = encoder.Decode(&ps)
	if err != nil {
		return nil
	}
	// loaded cache, but there were no contents, so return nil
	if len(ps) < 1 {
		return nil
	}

	return ps
}

// clearProcessCache removes the process cache file at the provided cache file
// path (cacheFp). If there is no cache file to remove, the
//
// An error is returned if there's an issue removing the cache file.
func clearProcessCache(cacheFp string) error {
	cacheLocation := filepath.Join(cacheFp, CacheFileName)
	err := os.Remove(cacheLocation)
	// if there's an error and it's not because the cache doesn't exist, return
	// an error.
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// encodeProcessCache takes processes and persists them to the local filesystem
// at the specified cacheFp. If the cacheFp does not exist, it attempts to
// create it. If a cache file already exists, encodeProcessCache will delete
// it.
//
// And error is returned when encodeProcessCache is unable to complete what is
// described above.
func encodeProcessCache(cacheFp string, ps Processes) error {
	// need to register ProcessStat as it'll come in as an interface within
	// Process.
	gob.Register(ProcessStat{})
	// if specified cache directory does not exist, create it.
	if _, err := os.Stat(cacheFp); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(cacheFp, 0666)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	cacheFileFp := filepath.Join(cacheFp, CacheFileName)
	cacheFile, err := os.Create(cacheFileFp)
	if err != nil {
		return err
	}
	defer cacheFile.Close()
	// persist the process details into cache

	encoder := gob.NewEncoder(cacheFile)
	err = encoder.Encode(ps)
	if err != nil {
		return err
	}

	return nil
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
	*FullPIDList = append(*FullPIDList, rootPID)
	parentID := pidlist[rootPID].ParentProcess
	if parentID > 1 {
		resolvePIDRelationship(FullPIDList, pidlist, parentID)
	}
	if parentID == 1 {
		*FullPIDList = append(*FullPIDList, 1)
	}
}

// GetProcessesByName looks up a process based on its name. In the case of
// linux, this is done by TODO(joshrosso). An error is returned if process
// lookup failed. If no process with the provided name is found, an empty slice
// is returned.
func GetProcessesByName(name string, opts ...ListOptions) ([]Process, error) {
	results := []Process{}
	ps, err := GetProcesses(opts...)
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

// getPIDs takes the procfs filepath and returns a list of all the known pids
// based on the directory name.
func getPIDsFromProcfs(procfsFp string) ([]int, error) {
	procDirs, err := os.ReadDir(procfsFp)
	if err != nil {
		return nil, err
	}

	pids := []int{}
	for _, p := range procDirs {
		// When a directory name is not [^0-9], its not a process and is skipped.
		pid, err := strconv.Atoi(p.Name())
		if err != nil {
			break
		}
		pids = append(pids, pid)
	}

	return pids, nil
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

func LoadProcessPath2(procfsFp string, pid int) (string, error) {
	exeLink, err := os.Readlink(filepath.Join(procfsFp, strconv.Itoa(pid), exeDir))
	if err != nil {
		return "", err
	}
	return exeLink, nil
}

// LoadProcessStat introspects the process's directory in procfs to retrieve
// relevant information and product a instance of Process. The generated
// Process object is then returned. No error is returned, as missing
// information or lack of access to data in procfs will result in missing
// information in the generated returned Process.
func LoadProcessStat(procfsFp string, pid int) Process {
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
			stat, err := os.ReadFile(filepath.Join(procfsFp, strconv.Itoa(pid), statDir))
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
	path, err := LoadProcessPath2(procfsFp, pid)
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
	stat := LoadStat2(procfsFp, pid)

	p := Process{
		ID:            pid,
		IsKernel:      isK,
		HasPermission: hasPerm,
		CommandName:   name,
		CommandPath:   path,
		ParentProcess: stat.ParentID,
		BinarySHA:     sha,
		OSSpecific:    &stat,
	}

	return p
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
		OSSpecific:    &stat,
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
func LoadStat2(procfsFp string, pid int) ProcessStat {
	ps := ProcessStat{}
	stat, err := os.ReadFile(filepath.Join(procfsFp, strconv.Itoa(pid), statDir))
	if err != nil {
		return ps
	}
	parsedStats := strings.Split(string(stat), " ")

	for i, stat := range parsedStats {
		//fmt.Printf("stat %d: %s\n", i, stat)
		switch i {
		case 0:
			ps.ID, _ = strconv.Atoi(stat)
		case 1:
			ps.FileName = stat
		case 2:
			ps.State = stat
		case 3:
			ps.ParentID, _ = strconv.Atoi(stat)
		case 4:
			ps.ProcessGroup, _ = strconv.Atoi(stat)
		case 5:
			ps.SessionID, _ = strconv.Atoi(stat)
		case 6:
			ps.TTY, _ = strconv.Atoi(stat)
		case 7:
			ps.TTYProcessGroup, _ = strconv.Atoi(stat)
		case 8:
			ps.TaskFlags = stat
		case 9:
			ps.MinorFaultQuantity, _ = strconv.Atoi(stat)
		case 10:
			ps.MinorFaultWithChildQuantity, _ = strconv.Atoi(stat)
		case 11:
			ps.MajorFaultQuantity, _ = strconv.Atoi(stat)
		case 12:
			ps.MajorFaultWithChildQuantity, _ = strconv.Atoi(stat)
		case 13:
			ps.UserModeTime, _ = strconv.Atoi(stat)
		case 14:
			ps.KernalTime, _ = strconv.Atoi(stat)
		case 15:
			ps.UserModeTimeWithChild, _ = strconv.Atoi(stat)
		case 16:
			ps.KernalTimeWithChild, _ = strconv.Atoi(stat)
		case 17:
			ps.Priority, _ = strconv.Atoi(stat)
		case 18:
			ps.Nice, _ = strconv.Atoi(stat)
		case 19:
			ps.ThreadQuantity, _ = strconv.Atoi(stat)
		case 20:
			ps.ItRealValue, _ = strconv.Atoi(stat)
		case 21:
			ps.StartTime, _ = strconv.Atoi(stat)
		case 22:
			ps.VirtualMemSize, _ = strconv.Atoi(stat)
		case 23:
			ps.ResidentSetMemSize, _ = strconv.Atoi(stat)
		case 24:
			ps.RSSByteLimit, _ = strconv.Atoi(stat)
		case 25:
			ps.StartCode = ConvertToHexMemoryAddress(stat)
		case 26:
			ps.EndCode = ConvertToHexMemoryAddress(stat)
		case 27:
			ps.StartStack = ConvertToHexMemoryAddress(stat)
		case 28:
			ps.ExtendedStackPointerAddress, _ = strconv.Atoi(stat)
		case 29:
			ps.ExtendedInstructionPointer, _ = strconv.Atoi(stat)
		case 30:
			ps.SignalPendingQuantity, _ = strconv.Atoi(stat)
		case 31:
			ps.SignalsBlockedQuantity, _ = strconv.Atoi(stat)
		case 32:
			ps.SignalsIgnoredQuantity, _ = strconv.Atoi(stat)
		case 33:
			ps.SiganlsCaughtQuantity, _ = strconv.Atoi(stat)
		case 34:
			ps.PlaceHolder1, _ = strconv.Atoi(stat)
		case 35:
			ps.PlaceHolder2, _ = strconv.Atoi(stat)
		case 36:
			ps.PlaceHolder3, _ = strconv.Atoi(stat)
		case 37:
			signalNumeric, _ := strconv.Atoi(stat)
			ps.ExitSignal = Signal(signalNumeric)
		case 38:
			ps.CPU, _ = strconv.Atoi(stat)
		case 39:
			ps.RealtimePriority, _ = strconv.Atoi(stat)
		case 40:
			ps.SchedulingPolicy, _ = strconv.Atoi(stat)
		case 41:
			ps.TimeSpentOnBlockIO, _ = strconv.Atoi(stat)
		case 42:
			ps.GuestTime, _ = strconv.Atoi(stat)
		case 43:
			ps.GuestTimeWithChild, _ = strconv.Atoi(stat)
		case 44:
			ps.StartDataAddress = ConvertToHexMemoryAddress(stat)
		case 45:
			ps.EndDataAddress = ConvertToHexMemoryAddress(stat)
		case 46:
			ps.HeapExpansionAddress = ConvertToHexMemoryAddress(stat)
		case 47:
			ps.StartCMDAddress = ConvertToHexMemoryAddress(stat)
		case 48:
			ps.EndCMDAddress = ConvertToHexMemoryAddress(stat)
		case 49:
			ps.StartEnvAddress = ConvertToHexMemoryAddress(stat)
		case 50:
			ps.EndEnvAddress = ConvertToHexMemoryAddress(stat)
		case 51:
			ps.ExitCode, _ = strconv.Atoi(stat)
		}
	}

	return ps
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

	for i, stat := range parsedStats {
		//fmt.Printf("stat %d: %s\n", i, stat)
		switch i {
		case 0:
			ps.ID, _ = strconv.Atoi(stat)
		case 1:
			ps.FileName = stat
		case 2:
			ps.State = stat
		case 3:
			ps.ParentID, _ = strconv.Atoi(stat)
		case 4:
			ps.ProcessGroup, _ = strconv.Atoi(stat)
		case 5:
			ps.SessionID, _ = strconv.Atoi(stat)
		case 6:
			ps.TTY, _ = strconv.Atoi(stat)
		case 7:
			ps.TTYProcessGroup, _ = strconv.Atoi(stat)
		case 8:
			ps.TaskFlags = stat
		case 9:
			ps.MinorFaultQuantity, _ = strconv.Atoi(stat)
		case 10:
			ps.MinorFaultWithChildQuantity, _ = strconv.Atoi(stat)
		case 11:
			ps.MajorFaultQuantity, _ = strconv.Atoi(stat)
		case 12:
			ps.MajorFaultWithChildQuantity, _ = strconv.Atoi(stat)
		case 13:
			ps.UserModeTime, _ = strconv.Atoi(stat)
		case 14:
			ps.KernalTime, _ = strconv.Atoi(stat)
		case 15:
			ps.UserModeTimeWithChild, _ = strconv.Atoi(stat)
		case 16:
			ps.KernalTimeWithChild, _ = strconv.Atoi(stat)
		case 17:
			ps.Priority, _ = strconv.Atoi(stat)
		case 18:
			ps.Nice, _ = strconv.Atoi(stat)
		case 19:
			ps.ThreadQuantity, _ = strconv.Atoi(stat)
		case 20:
			ps.ItRealValue, _ = strconv.Atoi(stat)
		case 21:
			ps.StartTime, _ = strconv.Atoi(stat)
		case 22:
			ps.VirtualMemSize, _ = strconv.Atoi(stat)
		case 23:
			ps.ResidentSetMemSize, _ = strconv.Atoi(stat)
		case 24:
			ps.RSSByteLimit, _ = strconv.Atoi(stat)
		case 25:
			ps.StartCode = ConvertToHexMemoryAddress(stat)
		case 26:
			ps.EndCode = ConvertToHexMemoryAddress(stat)
		case 27:
			ps.StartStack = ConvertToHexMemoryAddress(stat)
		case 28:
			ps.ExtendedStackPointerAddress, _ = strconv.Atoi(stat)
		case 29:
			ps.ExtendedInstructionPointer, _ = strconv.Atoi(stat)
		case 30:
			ps.SignalPendingQuantity, _ = strconv.Atoi(stat)
		case 31:
			ps.SignalsBlockedQuantity, _ = strconv.Atoi(stat)
		case 32:
			ps.SignalsIgnoredQuantity, _ = strconv.Atoi(stat)
		case 33:
			ps.SiganlsCaughtQuantity, _ = strconv.Atoi(stat)
		case 34:
			ps.PlaceHolder1, _ = strconv.Atoi(stat)
		case 35:
			ps.PlaceHolder2, _ = strconv.Atoi(stat)
		case 36:
			ps.PlaceHolder3, _ = strconv.Atoi(stat)
		case 37:
			signalNumeric, _ := strconv.Atoi(stat)
			ps.ExitSignal = Signal(signalNumeric)
		case 38:
			ps.CPU, _ = strconv.Atoi(stat)
		case 39:
			ps.RealtimePriority, _ = strconv.Atoi(stat)
		case 40:
			ps.SchedulingPolicy, _ = strconv.Atoi(stat)
		case 41:
			ps.TimeSpentOnBlockIO, _ = strconv.Atoi(stat)
		case 42:
			ps.GuestTime, _ = strconv.Atoi(stat)
		case 43:
			ps.GuestTimeWithChild, _ = strconv.Atoi(stat)
		case 44:
			ps.StartDataAddress = ConvertToHexMemoryAddress(stat)
		case 45:
			ps.EndDataAddress = ConvertToHexMemoryAddress(stat)
		case 46:
			ps.HeapExpansionAddress = ConvertToHexMemoryAddress(stat)
		case 47:
			ps.StartCMDAddress = ConvertToHexMemoryAddress(stat)
		case 48:
			ps.EndCMDAddress = ConvertToHexMemoryAddress(stat)
		case 49:
			ps.StartEnvAddress = ConvertToHexMemoryAddress(stat)
		case 50:
			ps.EndEnvAddress = ConvertToHexMemoryAddress(stat)
		case 51:
			ps.ExitCode, _ = strconv.Atoi(stat)
		}
	}

	return ps
}

// ConvertToHexMemoryAddress takes a memory address, represented as a decimal
// (the default for Linux's procfs) and converts it to a memory address in
// hexadecimal notation. Note the returned value will contain the '0x'
// notation.
func ConvertToHexMemoryAddress(decimalAddr string) string {
	// TODO(joshrosso): need to do something about error cases here
	d, _ := strconv.Atoi(decimalAddr)
	return fmt.Sprintf("0x%x", d)
}
