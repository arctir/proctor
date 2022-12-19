package cmd

import (
	"fmt"
	"os"
	"sort"

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
		// sort by number of commits
		sort.Sort(authors)
		out := newAuthorTableOutput(authors)
		output(out)
		return
	}

	out := newCommitTableOutput(commits, 30)
	output(out)
}

type ListOfAuthors []AuthorWrapper

func (a ListOfAuthors) Len() int           { return len(a) }
func (a ListOfAuthors) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ListOfAuthors) Less(i, j int) bool { return a[j].commitCount < a[i].commitCount }

// TODO(joshrosso)
func runDiffSource(cmd *cobra.Command, args []string) {
	return
}

type AuthorWrapper struct {
	commitCount int
	source.Person
}

// getAuthors takes a list of commits (c) and returns a slice of those authors.
func getAuthors(c []source.Commit) ListOfAuthors {
	authors := map[string]AuthorWrapper{}
	for _, commit := range c {
		if v, ok := authors[commit.Author.Email]; ok {
			authors[commit.Author.Email] = AuthorWrapper{v.commitCount + 1, commit.Author}
		} else {
			authors[commit.Author.Email] = AuthorWrapper{1, commit.Author}
		}
	}

	authorList := []AuthorWrapper{}
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
