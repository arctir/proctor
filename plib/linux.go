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

// newLinuxInspector takes an optional [LinuxInspectorConfig] and returns a
// configured LinuxInspector, which can be used to operator on processes with
// functions like [LinuxInspector.GetProcesses].
//
// The variadic nature of opts is only present to
// make this argument optional. Do not pass multiple opts arguments to this
// function. If you do, the last opt argument passed will be used.
//
// For any required confiuration not specified in the opts argument, including
// if opts is nil, defaults will be set.
func newLinuxInspector(opts ...InspectorConfig) (*LinuxInspector, error) {
	var config InspectorConfig
	// if opts was passed, used the last indexed argument
	if len(opts) > 0 {
		config = opts[len(opts)-1]
	}

	li := &LinuxInspector{
		InspectorConfig: config,
	}

	// set defaults to any empty fields that are required or have inadequate zero
	// values.
	setLinuxInspectorDefaults(li)

	return li, nil
}

// setLinuxInspectorDefaults takes a LinuxInspector instance and finds any required
// fields that aren't set to add defaults accordingly.
func setLinuxInspectorDefaults(li *LinuxInspector) {
	// if proc path isn't set, set it to /proc
	if li.LinuxConfig.ProcfsFilePath == "" {
		li.LinuxConfig.ProcfsFilePath = defaultProcDir
	}

	// if cache location isn't set it to ${XDG_DATA_HOME}/.local/share/proctor
	if li.CacheFilePath == "" {
		li.CacheFilePath = GetDefaultCacheLocation()
	}
}

// LoadProcesses gathers all the available process information using [procfs].
// Upon successful retreival, that data is stored in 2 places:
//
//  1. Within l.ps, which is an in-memory representation of processes stored within the struct.
//  2. Within the cache file, which is at the location specified by [InspectorConfig].
//
// When LoadProcesses is called, any process details store in l.ps are cleared
// and, assuming success, the existing cache file is replaced.
//
// [procfs]: https://en.wikipedia.org/wiki/Procfs
func (l *LinuxInspector) LoadProcesses() error {
	// Also reset the reference to in-memory cache of processes
	l.ps = Processes{}
	ps, err := getPIDsFromProcfs(l.LinuxConfig.ProcfsFilePath)
	if err != nil {
		return err
	}

	knownSHAs := map[string]string{}
	// for each pid, load its data
	for _, p := range ps {
		loadedProcess := LoadProcessStat(l.LinuxConfig.ProcfsFilePath, p, knownSHAs)
		// when the process is a kernel process and inspect is configured to not
		// include them, skip this process.
		if !l.LinuxConfig.IncludeKernel && loadedProcess.IsKernel {
			continue
		}
		// when the user doesn't have permission to access process details and the
		// inspector is configured to not include these, skip this process.
		if !l.LinuxConfig.IncludePermissionIssues && !loadedProcess.HasPermission {
			continue
		}

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

// ClearProcessCache will empty any cached process information stored from a
// LoadProcesses call. If called when there is no cache to clear, it returns
// without error. An error is returned if a cache is available to clear but
// it is unable to do so.
func (l *LinuxInspector) ClearProcessCache() error {
	err := clearProcessCache(l.CacheFilePath)
	if err != nil {
		return fmt.Errorf("failed to clear the existing process cache: %s", err)
	}
	return nil
}

// GetProcesses retrieves all process information available. If a cache
// pre-exists, GetProcesses will load the processes from that cache. If a cache
// does not exist, the implementation should run LoadProcesses, which loads
// process information from procfs and refreshes the cache. The cache may be
// ignored by settings the appropriate setting in [InspectorConfig].
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
			err := os.MkdirAll(cacheFp, 0777)
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

// GetProcessNameFromBinary attempts to resolve the executable name that
// created the process. This is resolved by looking up the symlink in:
//
//	/proc/${PID}/exec
//
// If the symlink cannot be resolved, an empty name and error is returned. This
// can commonly happen with kernel processes and processes which the user does
// not have permission to resolve the symlink. Upon error, it is up to the
// caller to determine if they'd like to resolve the name through other means.
func GetProcessNameFromBinary(procfsFp string, pid int) (string, error) {
	path, err := GetProcessPath(procfsFp, pid)
	if err != nil {
		return "", err
	}
	dirs := strings.Split(path, string(os.PathSeparator))
	if len(dirs) < 1 {
		return "", nil
	}
	return dirs[len(dirs)-1], nil
}

// NewSHAFromProcess takes a path to a file (likely a binary) and returns a
// SHA256 checksum representing its contents.
func NewSHAFromProcess(path string) string {
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

// GetProcessPath returns the path, or location, of the binary being executed
// as a process. To reliably determine the path, it reads the symbolic link in
// /proc/${PID}/exe and resolves the final file.
func GetProcessPath(procfsFp string, pid int) (string, error) {
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
// information in the generated returned Process. knownSHAs contains a map of
// SHA values where the key is the binary path. This enables lookup of already
// known SHAs without needing to rehash. To force rehash, use an empty map.
func LoadProcessStat(procfsFp string, pid int, knownSHAs map[string]string) Process {
	hasPerm := true
	isK := false
	var sha string
	name, err := GetProcessNameFromBinary(procfsFp, pid)

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
	path, err := GetProcessPath(procfsFp, pid)
	if err != nil {
		if os.IsPermission(err) {
			path = permDenied
			sha = permDenied
		} else {
			path = statError
			sha = statError
		}

	} else {
		// determine if sha is already known, if now, calculate it from file.
		if sum, ok := knownSHAs[path]; ok {
			sha = sum
		} else {
			sha = NewSHAFromProcess(path)
			knownSHAs[path] = sha
		}
	}
	stat := NewProcessStatFromFile(procfsFp, pid)

	p := Process{
		ID:            pid,
		IsKernel:      isK,
		HasPermission: hasPerm,
		CommandName:   name,
		CommandPath:   path,
		ParentProcess: stat.ParentID,
		BinarySHA:     sha,
		Type:          linuxProcessType,
		OSSpecific:    &stat,
	}

	return p
}

// NewProcessStatFromFile translates fields in the stat file
// (/proc/${PID}/stat) into structured data. See the [kernel docs] for a table
// of values found in a stat file.
//
// [kernel docs]: https://www.kernel.org/doc/html/latest/filesystems/proc.html#id10.
func NewProcessStatFromFile(procfsFp string, pid int) ProcessStat {
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

// ConvertToHexMemoryAddress takes a memory address, represented in [decimal
// notation] (base 10) (the default for Linux's procfs) and converts it to a
// memory address in [hexadecimal notation]. Note the returned value will contain
// the '0x' prefix.
//
// As an example, a valid decimal-represented memory address reported by procfs could be:
//
//	140732934197230
//
// When converted to hexadecimal notation, based on this function, this will be returned:
//
//	0x7ffef08d0fee
//
// For memory addresses, pblib should automatically do this
// translation to the hexadecimal notation held in a string. However, this
// fucntion is available in case you wish to do a conversion yourself.
//
// [decimal notation]: https://en.wikipedia.org/wiki/Decimal#Decimal_notation
// [hexadecimal notation]: https://en.wikipedia.org/wiki/Hexadecimal
func ConvertToHexMemoryAddress(decimalAddr string) string {
	d, _ := strconv.Atoi(decimalAddr)
	return fmt.Sprintf("0x%x", d)
}
