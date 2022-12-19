package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/arctir/proctor/plib"
	"github.com/arctir/proctor/source"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SetupCLI constructs the cobra hierachry to create the proctor CLI.
//
// Do not use this function in other Go pacakges. Instead, you should look to
// import the libraries used in the cmd packge directly. For example, [plib].
//
// [plib]: https://github.com/arctir/proctor/tree/main/plib
func SetupCLI() *cobra.Command {
	proctorCmd.AddCommand(processCmd)
	proctorCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(contribCmd)
	contribCmd.AddCommand(contribListCmd)
	processCmd.AddCommand(listCmd)
	processCmd.AddCommand(getCmd)
	processCmd.AddCommand(treeCmd)
	processCmd.AddCommand(fpCmd)

	return proctorCmd
}

// runProctor defines what should occur when `proctor ...` is run.
func runProctor(cmd *cobra.Command, args []string) {
	// if proctor is run without a command (argument), print help.
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}

// runProcess defines what should occur when `proctor process ...` is run.
func runProcess(cmd *cobra.Command, args []string) {
	// if proctor is run without a command (argument), print help.
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}

// runListProcesses defines the behavior of running:
// `proctor process ls ...`
func runListProcesses(cmd *cobra.Command, args []string) {
	opts := newProctorOptions(cmd.Flags())
	ps, err := createInspectorAndGetProcesses(opts)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("process collection failed: %s", err))
	}
	out, err := createListOutput(ps, opts)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed creating output for retrieved processes: %s", err))
	}

	output(out)
}

// runGetProcess defines the behavior of running:
// `proctor process get ...`
func runGetProcess(cmd *cobra.Command, args []string) {
	fs := cmd.Flags()
	opts := newProctorOptions(cmd.Flags())
	ps, err := createInspectorAndGetProcesses(opts)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("process collection failed: %s", err))
	}

	// use flags to determine how to resolve process(es)
	id, _ := fs.GetInt(idFlag)
	name, _ := fs.GetString(nameFlag)
	var out []byte
	switch {
	case id != 0:
		p := ps[id]
		out, err = createSingleOutput(p, opts)
		if err != nil {
			outputErrorAndFail(fmt.Sprintf("failed creating output for process: %s", err))
		}
	case name != "":
		matchedPs := findAllProcessesWithName(name, ps)
		out, err = createListOutput(matchedPs, opts)
		if err != nil {
			outputErrorAndFail(fmt.Sprintf("failed creating output for processes: %s", err))
		}
	default:
		cmd.Help()
	}

	output(out)
}

// runTreeProcess defines the behavior of running:
// `proctor process tree ...`
func runTreeProcess(cmd *cobra.Command, args []string) {
	pid, err := parseID(args)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("please pass a valid pid (int); we received: %s", args[0]))
	}
	opts := newProctorOptions(cmd.Flags())
	ps, err := createInspectorAndGetProcesses(opts)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("process collection failed: %s", err))
	}
	if ps[pid] == nil {
		outputErrorAndFail(fmt.Sprintf("failed to find process with id: %d", pid))
	}

	// collect all processes from the specified and recursively to every parent.
	relatedPs := []plib.Process{}
	relatedPs = append(relatedPs, *ps[pid])
	currentParentPid := ps[pid].ParentProcess
	for {
		// we've reached the root (likely the init system).
		if currentParentPid == 0 {
			break
		}
		// if we can't resolve details about the parent process, stop gathering the
		// hierarchy.
		if ps[currentParentPid] == nil {
			break
		}
		relatedPs = append(relatedPs, *ps[currentParentPid])
		currentParentPid = ps[currentParentPid].ParentProcess
	}

	o, err := createSliceListOutput(relatedPs, opts)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed creating output for processes: %s", err))
	}
	output(o)
}

// runFingerPrintProcess defines the behavior for running:
// `proctor process finger-print ...`
func runFingerPrintProcess(cmd *cobra.Command, args []string) {
	pid, err := parseID(args)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("please pass a valid pid (int); we received: %s", args[0]))
	}
	opts := newProctorOptions(cmd.Flags())
	ps, err := createInspectorAndGetProcesses(opts)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("process collection failed: %s", err))
	}
	if ps[pid] == nil {
		outputErrorAndFail(fmt.Sprintf("failed to find process with id: %d", pid))
	}

	if ps[pid].BinarySHA == "" {
		outputErrorAndFail(fmt.Sprintf("process %d is missing details about its binary binary checksum.", pid))
	}
	combinedHashes := ps[pid].BinarySHA
	// collect all processes from the specified and recursively to every parent.
	currentParentPid := ps[pid].ParentProcess
	for {
		// we've reached the root (likely the init system).
		if currentParentPid == 0 {
			break
		}
		// if we can't resolve details about the parent process, there may be an
		// issue with permission and the finger print will not be valid.
		if ps[currentParentPid] == nil {
			outputErrorAndFail(fmt.Sprintf("could not gather details on parent process: %d and thus could not generate a finger print. error: %s", currentParentPid, err))
		}
		combinedHashes += ps[currentParentPid].BinarySHA
		currentParentPid = ps[currentParentPid].ParentProcess
	}

	fp := sha256.Sum256([]byte(combinedHashes))
	output([]byte(hex.EncodeToString(fp[:])))
}

// parseID is a helper function to determine if the first argument passed to
// the command is a valid ID (int).
func parseID(args []string) (int, error) {
	// user must specify an ID
	if len(args) < 1 {
		return 0, fmt.Errorf("please provide a pid (int)")
	}
	pid, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// createInspectorAndGetProcesses is a helper function since most CLI commands will need table:
// 1. Create a new LinuxInspector
// 2. Setup configuration
// 3. Retrieve a list of processes
func createInspectorAndGetProcesses(opts proctorOpts) (plib.Processes, error) {
	conf := plib.InspectorConfig{
		LinuxConfig: plib.LinuxInspectorConfig{
			IncludeKernel:           opts.includeKernel,
			IncludePermissionIssues: opts.includePermIssue,
		},
	}
	insp, err := plib.NewInspector(conf)
	if err != nil {
		return nil, fmt.Errorf("failed setting up library to retrieve processes: %s", err)
	}
	// if reset cache was set, clear the cache before attempting to load processes
	if opts.resetCache {
		insp.ClearProcessCache()
	}
	ps, err := insp.GetProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed retrieving processes via Linux APIs: %s", err)
	}
	return ps, nil
}

// findAllProcessesWithName looks through all processes (ps) and find any
// process where the [plib.Process]'s CommandName is equal to the provided
// name. Since there can be multiple processes with the same command name, this
// returns another processes (map/list).
func findAllProcessesWithName(name string, ps plib.Processes) plib.Processes {
	matchedPs := plib.Processes{}
	for _, p := range ps {
		if p.CommandName == name {
			matchedPs[p.ID] = p
		}
	}

	return matchedPs
}

func output(out []byte) {
	fmt.Printf("%s", out)
}

func outputErrorAndFail(msg string) {
	fmt.Println(msg)
	// exit(1) is the catchall for general errors.
	os.Exit(1)
}

func createSingleOutput(ps *plib.Process, opts proctorOpts) ([]byte, error) {
	var out []byte
	switch opts.outType {
	case jsonOut:
		out = createJSONSingleOutput(ps)
	default:
		out = createTableSingleOutput(ps)
	}

	return out, nil
}

func createSliceListOutput(ps []plib.Process, opts proctorOpts) ([]byte, error) {
	var out []byte
	switch opts.outType {
	case jsonOut:
		out = createJSONSliceListOutput(ps)
	default:
		out = createTableSliceListOutput(ps)
	}

	return out, nil
}

func createListOutput(ps plib.Processes, opts proctorOpts) ([]byte, error) {
	var out []byte
	switch opts.outType {
	case jsonOut:
		out = createJSONListOutput(ps)
	default:
		out = createTableListOutput(ps)
	}

	return out, nil
}

func createJSONSliceListOutput(ps []plib.Process) []byte {
	out, _ := json.Marshal(ps)
	return out
}

func createJSONListOutput(ps plib.Processes) []byte {
	out, _ := json.Marshal(ps)
	return out
}

func createJSONSingleOutput(ps *plib.Process) []byte {
	out, _ := json.Marshal(ps)
	return out
}

func createTableSingleOutput(p *plib.Process) []byte {
	if p == nil {
		return []byte{}
	}

	psToReturn := []string{
		strconv.Itoa(p.ID),
		p.CommandName,
		p.CommandPath,
		p.BinarySHA,
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"PID", "name", "location", "SHA"})
	table.Append(psToReturn)
	table.Render()
	return buf.Bytes()
}

// newCommitTableOutput takes a list of commits and create a table output
// represented in bytes. It offers a lengthLimit argument which allows
// limitting the amount of bytes used when printing in the table.
func newCommitTableOutput(commits []source.Commit, lengthLimit int) []byte {
	listOfCommits := [][]string{}
	for _, c := range commits {
		truncatedMsg := []byte{}
		if len(c.Message) > lengthLimit {
			truncatedMsg = c.Message[:lengthLimit]
		} else {
			truncatedMsg = c.Message
		}
		truncatedAuthor := []byte(c.Author.Email)
		if len(truncatedAuthor) > lengthLimit {
			truncatedAuthor = truncatedAuthor[:lengthLimit]
		}
		finalCommitMsg := strings.ReplaceAll(string(truncatedMsg), "\n", " ")
		listOfCommits = append(listOfCommits, []string{
			c.Hash.String(),
			finalCommitMsg,
			string(truncatedAuthor),
		})
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"SHA", "Message", "Author"})
	table.AppendBulk(listOfCommits)
	table.Render()
	return buf.Bytes()
}

func newAuthorTableOutput(authors []AuthorWrapper) []byte {
	listOfAuthors := [][]string{}
	for _, a := range authors {
		listOfAuthors = append(listOfAuthors, []string{
			strconv.Itoa(a.commitCount),
			a.Name,
			a.Email,
		})
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"Commits", "Name", "Email"})
	table.AppendBulk(listOfAuthors)
	table.SetAutoWrapText(false)
	table.Render()
	return buf.Bytes()
}

func createTableListOutput(ps plib.Processes) []byte {
	listOfPs := [][]string{}
	for _, p := range ps {
		listOfPs = append(listOfPs, []string{
			strconv.Itoa(p.ID),
			p.CommandName,
			p.CommandPath,
			p.BinarySHA,
		})
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"PID", "name", "location", "SHA"})
	table.AppendBulk(listOfPs)
	table.Render()
	return buf.Bytes()
}

func createTableSliceListOutput(ps []plib.Process) []byte {
	listOfPs := [][]string{}
	for _, p := range ps {
		listOfPs = append(listOfPs, []string{
			strconv.Itoa(p.ID),
			p.CommandName,
			p.CommandPath,
			p.BinarySHA,
		})
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader([]string{"PID", "name", "location", "SHA"})
	table.AppendBulk(listOfPs)
	table.Render()
	return buf.Bytes()
}

// sourceOpts provides details on how source-related details should be
// retrieved
type sourceOpts struct {
	retrieveOnlyAuthors bool
	// used when you want to limit commit retrieval to a single tag
	singleTag string
	// used when you want to compare commits between 2 tags, required tagTwo to
	// be set.
	tagOne string
	// used when you want to compare commits between 2 tags, required tagOne to
	// be set.
	tagTwo string
}

func newSourceOptions(fs *pflag.FlagSet) sourceOpts {
	roa, _ := fs.GetBool(authorsFlag)
	singleTag, _ := fs.GetString(tagFlag)

	return sourceOpts{
		retrieveOnlyAuthors: roa,
		singleTag:           singleTag,
		tagOne:              "",
		tagTwo:              "",
	}
}

func newProctorOptions(fs *pflag.FlagSet) proctorOpts {
	ot := resolveOutputType(fs)
	fko, _ := fs.GetBool(includeKernelFlag)
	ipi, _ := fs.GetBool(includePermIssueFlag)
	rc, _ := fs.GetBool(resetCacheFlag)

	return proctorOpts{
		outType:          ot,
		includeKernel:    fko,
		includePermIssue: ipi,
		resetCache:       rc,
	}
}

func resolveOutputType(fs *pflag.FlagSet) outputType {
	of, err := fs.GetString(outputFlag)
	// default if there are ever issues finding flag
	if err != nil {
		return tableOut
	}
	switch of {
	case "json":
		return jsonOut
	case "table":
		return tableOut
	}

	// default OutputType
	return tableOut
}
