package plib

// LinuxInspector is the Linux implementation of inspector for loading,
// caching, and returning all process information available. It does this by
// reading the [process virtual filesystem] (procfs) and getting information
// from various files, such as stat.
//
// It is not recommended you a LinuxInspector directly, instead use the
// [NewInspector] constructor which will ensure configuration and defaults are
// respected.
//
// [process virtual filesystem]: https://docs.kernel.org/filesystems/proc.html
type LinuxInspector struct {
	InspectorConfig
	ps Processes
}

// LinuxInspectorConfig can be used to set Linux-specific settings when
// creating an inspector. This config should be embedded into a
// [InspectorConfig] struct as the OSSpecificConfig value.
type LinuxInspectorConfig struct {
	// The location on the filesystem where the virtual filesystem (procfs) can
	// be found. By default, this is set to /proc. Unless you are writing tests
	// or operating a very-customized Linux box, you probably should not set
	// this.
	ProcfsFilePath string
}

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
	// A character that reflects:SchedulingPolicy the state of the process.
	// R: The process is currently running.
	// S: The process is sleeping (waiting for an event to occur).
	// D: The process is in uninterruptible sleep (usually waiting for I/O to complete).
	// Z: The process is a zombie (a terminated process that has not been cleaned up by its parent).
	// T: The process is stopped (either by a job control signal or by the trace system call).
	// t: The process is traced (being debugged by another process).
	// X: The process is dead (should never be seen).
	// x: The process is dead (should never be seen).
	// K: The process is Wakekill (should never be seen).
	// W: The process is Waking (should never be seen).
	// P: The process is Parked (should never be seen).
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
	// The time (number of ticks) that occurred between boot and when the process
	// was assigned its PID. This value is typically measured in clock ticks and
	// is expressed as the number of seconds that have elapsed since the system
	// was booted
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
