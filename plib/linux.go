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

type Signal int

const (
	SIGHUP Signal = iota
	SIGINT
	SIGQUIT
	SIGILL
	SIGTRAP
	SIGABRT
	SIGIOT
	SIGBUS
	SIGFPE
	SIGKILL
	SIGUSR1
	SIGSEGV
	SIGUSR2
	SIGPIPE
	SIGALRM
	SIGTERM
	SIGSTKFLT
	SIGCHLD
	SIGCONT
	SIGSTOP
	SIGTSTP
	SIGTTIN
	SIGTTOU
	SIGURG
	SIGXCPU
	SIGXFSZ
	SIGVTALRM
	SIGPROF
	SIGWINCH
	SIGIO
	SIGPWR
	SIGSYS
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
	// The process's ID
	// Also known as pid.
	ID int
	// The filename of the executable associated with the process. There is no
	// guarantee this actually reflects the filename of the original executable
	// as this value can be set using the system call
	// [prctl](https://man7.org/linux/man-pages/man2/prctl.2.html).
	// Also known as tcomm
	FileName string
	// A character that reflects the state of the process.
	// R: The process is currently running.
	//S: The process is sleeping (waiting for an event to occur).
	//D: The process is in uninterruptible sleep (usually waiting for I/O to complete).
	//Z: The process is a zombie (a terminated process that has not been cleaned up by its parent).
	//T: The process is stopped (either by a job control signal or by the trace system call).
	//t: The process is traced (being debugged by another process).
	//X: The process is dead (should never be seen).
	//x: The process is dead (should never be seen).
	//K: The process is Wakekill (should never be seen).
	//W: The process is Waking (should never be seen).
	//P: The process is Parked (should never be seen).
	State string
	// The parent process's ParentID ( or the unique ID of the process that created this process ).
	// Also known as ppid.
	ParentID int
	// The process group this process belongs to. This value reflects that process group's ID.
	// Also known as ppid.
	ProcessGroup int
	// The ID of the Linus session this process belongs to.
	// Also known as sid.
	SessionID int // sid
	// The TTY number the process is associated with. When this value is 0, the
	// process is not associated with a TTY (and thus not connected to a user).
	// Also known as tty_grp.
	TTY int
	// Also known as tty_pgrp.
	TTYProcessGroup int
	// the task's current scheduling flags which are expressed in hexadecimal
	// notation and with zeros suppressed.
	TaskFlags string
	// A page fault is an event that occurs when a process tries to access a page
	// in virtual memory that is not currently present in physical memory. A
	// minor page fault occurs when the page being accessed has been paged out to
	// disk, but can be quickly retrieved from the disk (thanks to cache) and
	// returned to memory.
	MinorFaultQuantity int
	// Quantity of time child process of the current process tries to access a
	// page that can be quickly retrieved from disk (thanks to cache).
	MinorFaultWithChildQuantity int
	// A page fault is an event that occurs when a process tries to access a page
	// in virtual memory that is not currently present in physical memory. A
	// major page fault occurs when the page being accessed must be retrieved
	// from disk and is not currently in the page cache. This type of page fault
	// is more costly because it involves reading the page from disk, which is
	// much slower than accessing memory. As a result, major page faults can
	// significantly impact the performance of a process.
	MajorFaultQuantity int
	// Quantity of time child process of the current process tries to access a
	// page that must be retrieved from disk.
	MajorFaultWithChildQuantity int
	// The amount of time that the process has spent in user mode (userland).
	// The amount is measured in jiffies
	// Also known as `utime`.
	UserModeTime int
	// The amount of time that the process has spent in kernel mode.
	// Also known as `stime`.
	KernalTime int
	// The amount of time the process's children have spent in user mode (userland).
	// Also known as `cutime`.
	UserModeTimeWithChild int
	// The amount of time the process's children have spent in kernel mode.
	// Also known as `cstime`.
	KernalTimeWithChild int
	// A value used by the kernel to prioritize CPU and other resources relative
	// to other process's. This value can change over time. To view a process's
	// initial value, see Priority. Lower numerical values correspond to higher
	// priority levels, and higher numerical values correspond to lower priority
	// levels.
	Priority int
	// A value used by the kernel to prioritize CPU and other resources relative
	// to other process's. This can be thought of as the initial priority value
	// and is static. Lower numerical values correspond to higher priority
	// levels, and higher numerical values correspond to lower priority levels.
	Nice int
	// The number of threads that are associated with the process. Process's may
	// create multiple threads to do concurrent processing within the same
	// address space (or space of shared resources).
	// Also known as `num_threads`.
	ThreadQuantity int
	// This field was used in older versions of Linux to represent the interval
	// timer value of a process, but it has since been replaced by the
	// it_virt_value and it_prof_value fields, which provide more accurate and
	// up-to-date information about the interval timer.
	// Also known as `it_real_value`.
	ItRealValue int
	// The time (number of ticks) that occurred between boot and when the process was assigned its PID.
	// This value is typically measured in clock ticks and is expressed as the
	// number of seconds that have elapsed since the system was booted
	StartTime int
	// The virtual memory size of the process. This value is typically measured
	// in bytes, and can be used to determine the amount of memory that the
	// process has available for its own use (not necessarily what it is using).
	// Also known as vsize.
	VirtualMemSize int
	// The amount of memory that is currently being used by the process and is
	// resident in physical memory, and can be used to determine the process's
	// memory usage. This value is typically measured in bytes.
	// Also known as rss.
	ResidentSetMemSize int
	// The maximum amount of memory that the process is allowed to use, and is
	// used to prevent the process from consuming too much memory and affecting
	// the performance of other processes on the system. Note that resident set
	// size represents actual memory, where as things like the virtual memory
	// space are an abstraction that can tie the concept of memory to things like
	// "disk", ie the paging out of memory.
	// Also known as rsslim.
	RSSByteLimit int
	// This value is a hexadecimal representation of the starting address of the
	// code segment, and can be used to determine the location of the process's
	// code in its virtual address space The code segment can be thought of as
	// where the process's instruction set is stored (e.g. functions, routines,
	// etc).
	StartCode string
	// This value is a hexadecimal representation of the ending address of the code
	// segment, and can be used to determine the location of the process's The
	// code segment can be thought of as where the process's instruction set is
	// stored (e.g. functions, routines, etc).
	EndCode string
	// This value is a hexadecimal representation of the ending address of the code
	// segment, and can be used to determine the location of the process's The
	// code segment can be thought of as where the process's instruction set is
	// stored (e.g. functions, routines, etc).
	StartStack string
	// The memory address of the register that holds the current value of the
	// stack pointer for the process. The stack pointer points to the top of the
	// stack for the process, and is used by the processor to access memory on
	// the stack.
	// Also known as esp.
	ExtendedStackPointerAddress int
	// The memory address that points to the next instruction to be executed by
	// the CPU.
	// Also known as eip.
	ExtendedInstructionPointer int
	// The number signals (e.g. SIGINT) the process currently has pending.
	// Also known as pending
	SignalPendingQuantity int
	// The number signals (e.g. SIGINT) the process has blocked. Processes may be
	// able to block signals if they have run the system call
	// [sigprocmask()](https://man7.org/linux/man-pages/man2/sigprocmask.2.html),
	// often used so they can do a critical opration without being interupted by
	// a signal.
	// Also known as blocked
	SignalsBlockedQuantity int
	// The number of signals ignored by the process, which can be achieved when a
	// process runs the
	// [signal()](https://man7.org/linux/man-pages/man2/signal.2.html) system
	// call where the disposition is set to SIG_IGN.
	// Also known as sigign
	SignalsIgnoredQuantity int
	// The number of signals caught by the process, which is is handling.
	// Processes may handle signals by calling the signal() system call.
	// Also known as sigcatch
	SiganlsCaughtQuantity int
	// This used to be the wchan address, but that has been moved to
	// /proc/PID/wchan.
	// Also known as 0.
	PlaceHolder1 int
	// Also known as 0
	PlaceHolder2 int
	// Also known as 0
	PlaceHolder3 int
	// The signal value that caused the process to exit. For example, a value of
	// 9 would correspond to SIGKILL.
	// TODO(joshrosso): Create an enum translating these signals.
	ExitSignal Signal
	// The (v)CPU a task is scheduled on. For example 0 represents the first
	// core, and 4 represents the 5th.
	// Also known as task_cpu.
	CPU int
	// The real time priority of a process, used for making scheduling decisions.
	// A higher value represents higher priority. Real-time scheduling is a
	// method of scheduling processes on a computer in which each process is
	// assigned a priority level, and processes with higher priority levels are
	// given more CPU time than processes with lower priority levels.
	// Also known as rt_priority
	RealtimePriority int
	// Represents the scheduling policy of the process.
	// 0: SCHED_NORMAL
	// 1: SCHED_FIFO
	// 2: SCHED_RR
	// 3: SCHED_BATCH
	// Also known as policy
	SchedulingPolicy int
	// The amount of time a process waited for block read/write operations.
	// There is some nuance to how this value is calculated and should not be
	// used as an accurate performance metric.
	// Also known as blkio_ticks.
	TimeSpentOnBlockIO int
	// The amount of time the process spent running in a (guest) virtual machine).
	// Also known as gtime.
	GuestTime int
	// The amount of time the process's children have spent running in a virtual
	// machine.
	// Also known as cgtime.
	GuestTimeWithChild int
	// The beginning (virtual) memory address of the data segement for the
	// process. This can include things like data structures or heap.
	// Also known as start_data.
	StartDataAddress string
	// The ending (virtual) memory address of the data segement for the
	// process. This can include things like data structures or heap.
	// Also known as end_data.
	EndDataAddress string
	// The starting (virtual) memory address of the heap. This should be within
	// the StartDataAddress.
	// Also known as start_brk
	HeapExpansionAddress string
	// The starting memory address for the block of memory containing the command
	// line arguments used to run the executable, thus creating the process.
	// Also known as arg_start.
	StartCMDAddress string
	// The ending memory address for the block of memory containing the command
	// line arguments used to run the executable, thus creating the process.
	// Also known as arg_end.
	EndCMDAddress string
	// The starting memory address for the block of memory that contains the
	// environment variables set by both the system and user, which are available
	// to the process.
	// Also known as env_start.
	StartEnvAddress string
	// The ending memory address for the block of memory that contains the
	// environment variables set by both the system and user, which are available
	// to the process.
	// Also known as env_end.
	EndEnvAddress string
	// The exit code that is reported to the parent process based on this process
	// ending. Only relevant if a process has exited.
	ExitCode int
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
	*FullPIDList = append(*FullPIDList, rootPID)
	parentID := pidlist[rootPID].ParentProcess
	if parentID > 1 {
		resolvePIDRelationship(FullPIDList, pidlist, parentID)
	}
	if parentID == 1 {
		*FullPIDList = append(*FullPIDList, 1)
	}
}

func RunGetProcessForRelationship(name string, opts ...ListOptions) ProcessRelation {
	ps, err := GetProcesses(opts...)
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
			break
		}

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
			ps.ExitSignal, _ = strconv.Atoi(stat)
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
// hexadecimal notation. Note the returned value will contain the '0x' hex
// prefix.
func ConvertToHexMemoryAddress(decimalAddr string) string {
	// TODO(joshrosso): need to do something about error cases here
	d, _ := strconv.Atoi(decimalAddr)
	return fmt.Sprintf("0x%x", d)
}
