package cmd

import (
	"fmt"
	"os"

	"github.com/arctir/proctor/source"
	"github.com/spf13/cobra"
)

// runSource defines what should occur when `proctor source ...` is run.
func runSource(cmd *cobra.Command, args []string) {
	// if proctor is run without a command (argument), print help.
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}

// runContrib defines what should occur when `proctor source contrib ...` is
// run.
func runContrib(cmd *cobra.Command, args []string) {
	// if proctor is run without a command (argument), print help.
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}

// runContribSource defines the behavior of running:
// `proctor process ls ...`
func runContribList(cmd *cobra.Command, args []string) {
	opts := newSourceOptions(cmd.Flags())
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
	commits, err := getCommits(args[0])
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed resolving commits, underlying error: %s", err))
	}

	// when --authors is specified, create an output that exclusively contains
	// authors.
	if opts.retrieveOnlyAuthors {
		authors := getAuthors(commits)
		out := newAuthorTableOutput(authors)
		output(out)
		return
	}

	out := newCommitTableOutput(commits, 30)
	output(out)
}

// TODO(joshrosso)
func runDiffSource(cmd *cobra.Command, args []string) {
	return
}

// getAuthors takes a list of commits (c) and returns a slice of those authors.
func getAuthors(c []source.Commit) []source.Person {
	authors := map[string]source.Person{}
	for _, commit := range c {
		authors[commit.Author.Email] = commit.Author
	}

	authorList := []source.Person{}
	for _, v := range authors {
		authorList = append(authorList, v)
	}

	return authorList
}

// getCommits is a healper function that returns all the commits for a
// repostiory, passed as url.
func getCommits(url string) ([]source.Commit, error) {
	repo, err := source.NewInMemRepo(url)
	if err != nil {
		return nil, err
	}

	gm := source.NewGitManager()
	commits, err := gm.GetCommits(*repo)
	if err != nil {
		return nil, err
	}
	return commits, nil
}
