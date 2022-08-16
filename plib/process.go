package plib

type Inspector interface {
  ListProcesses() []Process
  GetProcess(qo ProcessQueryOptions) Process
}

type ProcessQueryOptions struct {
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
	Stat          any
}
