package cmd

type outputType int

const (
	jsonOut outputType = iota
	tableOut
)

const (
	outputFlag           = "output"
	authorsFlag          = "authors"
	tagFlag              = "tag"
	tagOneFlag           = "tag1"
	tagTwoFlag           = "tag2"
	includeKernelFlag    = "include-kernel"
	includePermIssueFlag = "include-permission-issues"
	resetCacheFlag       = "reset-cache"
	nameFlag             = "name"
	idFlag               = "id"
)

type proctorOpts struct {
	outType          outputType
	includeKernel    bool
	includePermIssue bool
	resetCache       bool
}

// CLI flags to intialize
func init() {
	// output
	getCmd.Flags().StringP(outputFlag, "o", "table", "Output type for command [table (default), json].")
	listCmd.Flags().StringP(outputFlag, "o", "table", "Output type for command [table (default), json].")
	treeCmd.Flags().StringP(outputFlag, "o", "table", "Output type for command [table (default), json].")

	// cache-reset
	listCmd.Flags().Bool(resetCacheFlag, false, "Refreshs the cache, making this call read all its data from the OS, replenish the cache, and output.")
	getCmd.Flags().Bool(resetCacheFlag, false, "Refreshs the cache, making this call read all its data from the OS, replenish the cache, and output.")
	treeCmd.Flags().Bool(resetCacheFlag, false, "Refreshs the cache, making this call read all its data from the OS, replenish the cache, and output.")
	fpCmd.Flags().Bool(resetCacheFlag, false, "Refreshs the cache, making this call read all its data from the OS, replenish the cache, and output.")

	// kernel filter
	getCmd.Flags().Bool(includeKernelFlag, false, "Include kernel processes in out, default is false.")
	listCmd.Flags().Bool(includeKernelFlag, false, "Include kernel processes in out, default is false.")
	treeCmd.Flags().Bool(includeKernelFlag, false, "Include kernel processes in out, default is false.")

	// permission filter
	listCmd.Flags().Bool(includePermIssueFlag, false, "Include processes that proctor failed to introspect due to permission issues.")
	treeCmd.Flags().Bool(includePermIssueFlag, false, "Include processes that proctor failed to introspect due to permission issues.")
	getCmd.Flags().Bool(includePermIssueFlag, false, "Include processes that proctor failed to introspect due to permission issues.")

	// get flags
	getCmd.Flags().String(nameFlag, "", "Get processes by the name. This will return a list of processes since processes may share the same command name.")
	getCmd.Flags().Int(idFlag, 0, "Get processes ID. This returns a single process since IDs are unique to processes")

	// contrib flags
	contribCmd.Flags().Bool(authorsFlag, false, "Limit output to details about contributing authors.")
	contribCmd.Flags().StringP(tagFlag, "t", "", "Limit the results to a single tag.")
	listCmd.Flags().String(tagOneFlag, "", "Output type for command [table (default), json].")
	listCmd.Flags().String(tagTwoFlag, "", "Output type for command [table (default), json].")
}
