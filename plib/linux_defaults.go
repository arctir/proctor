package plib

import "os"

type Signal int

const (
	defaultProcDir = string(os.PathSeparator) + "proc"
	cmdDir         = "cmdline"
	statDir        = "stat"
	exeDir         = "exe"
	nullCharacter  = "\x00"
	permDenied     = "PERM_DENIED"
	statError      = "ERROR_READING_STAT"
)

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
