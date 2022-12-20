package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/arctir/proctor/platforms/github"
	"github.com/arctir/proctor/source"
	"github.com/spf13/cobra"
)

// authorWrapper contains a source.Person and holds a commmit count the user
// may wish to change over time.
type authorWrapper struct {
	commitCount int
	source.Person
}

// authorWrappers exists purely to implement sort.Interface.
type authorWrappers []authorWrapper

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

// runArtifacts defines what should occur when `proctor source
// artifacts ...` is run.
func runArtifacts(cmd *cobra.Command, args []string) {
	// if proctor is run without a command (argument), print help.
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
}

// runListArtifacts defines what should occur when `proctor source
// artifacts list ...` is run.
func runListArtifacts(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
	gh := github.NewGHManager()
	repo := strings.Split(args[0], "https://github.com/")
	if len(repo) < 2 {
		outputErrorAndFail(fmt.Sprintf("repository (%s) provided was invalid. At this time we only support https://github.com/$ORG/$REPO.", args[0]))
	}
	orgAndRepo := repo[1]
	arts, err := gh.GetArtifacts(orgAndRepo)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed retrieving artifacts: %s", err))
	}
	out := newArtifactListTableOutput(arts)
	output(out)
}

// runGetArtifacts defines what should occur when `proctor source
// artifacts get ...` is run.
func runGetArtifacts(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
	opts := newSourceOptions(cmd.Flags())
	if opts.singleTag == "" {
		outputErrorAndFail("please specify --tag when looking up artifacts")
	}

	gh := github.NewGHManager()
	repo := strings.Split(args[0], "https://github.com/")
	if len(repo) < 2 {
		outputErrorAndFail(fmt.Sprintf("repository (%s) provided was invalid. At this time we only support https://github.com/$ORG/$REPO.", args[0]))
	}
	orgAndRepo := repo[1]
	releases, err := gh.GetArtifacts(orgAndRepo)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed retrieving artifacts: %s", err))
	}
	arts := []github.Artifact{}
	for _, r := range releases {
		if r.Tag == opts.singleTag {
			arts = r.Artifacts
		}
	}
	if len(arts) < 1 {
		outputErrorAndFail(fmt.Sprintf("failed to find any artifacts for tag (%s)", opts.singleTag))
	}
	out := newArtifactGetTableOutput(arts)
	output(out)
}

// runContribSource defines the behavior of running:
// `proctor process ls ...`
func runContribList(cmd *cobra.Command, args []string) {
	opts := newSourceOptions(cmd.Flags())
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}

	commits := []source.Commit{}
	var err error
	if opts.singleTag != "" {
		commits, err = getCommitsForTag(args[0], opts.singleTag)
		if err != nil {
			outputErrorAndFail(fmt.Sprintf("failed resolving commits, underlying error: %s", err))
		}
	} else {
		commits, err = getCommits(args[0])
		if err != nil {
			outputErrorAndFail(fmt.Sprintf("failed resolving commits, underlying error: %s", err))
		}
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

// runDiffSource is the equivelent to `proctor source contrib diff ...`.
func runDiffSource(cmd *cobra.Command, args []string) {
	opts := newSourceOptions(cmd.Flags())
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
	if opts.tagOne == "" {
		outputErrorAndFail("please provide value for --tag1")
	}
	if opts.tagTwo == "" {
		outputErrorAndFail("please provide value for --tag2")
	}

	var err error
	commits1, err := getCommitsForTag(args[0], opts.tagOne)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed resolving commits, underlying error: %s", err))
	}
	commits2, err := getCommitsForTag(args[0], opts.tagTwo)
	if err != nil {
		outputErrorAndFail(fmt.Sprintf("failed resolving commits, underlying error: %s", err))
	}
	reverseCommitsOrder(commits1)
	reverseCommitsOrder(commits2)

	commitsOnlyInOne := []source.Commit{}
	// detect commits only in 1
	for i := range commits1 {
		if len(commits2)-1 < i {
			commitsOnlyInOne = append(commitsOnlyInOne, commits1[i])
			continue
		}
		if commits2[i].Hash != commits1[i].Hash {
			commitsOnlyInOne = append(commitsOnlyInOne, commits1[i])
		}
	}
	reverseCommitsOrder(commitsOnlyInOne)

	commitsOnlyInTwo := []source.Commit{}
	// detect commits only in 2
	for i := range commits2 {
		if len(commits1)-1 < i {
			commitsOnlyInTwo = append(commitsOnlyInTwo, commits2[i])
			continue
		}
		if commits2[i].Hash != commits1[i].Hash {
			commitsOnlyInTwo = append(commitsOnlyInTwo, commits2[i])
		}
	}
	reverseCommitsOrder(commitsOnlyInTwo)

	// when --authors is specified, create an output that exclusively contains
	// authors.
	if opts.retrieveOnlyAuthors {
		authors := getAuthors(commitsOnlyInOne)
		// sort by number of commits
		sort.Sort(authors)
		out := newAuthorTableOutput(authors)
		output(out)
		return
	}

	out := newCommitDiffTableOutput(commitsOnlyInOne, opts.tagOne, commitsOnlyInTwo, opts.tagTwo, 30)
	output(out)
}

func reverseCommitsOrder(commits []source.Commit) {
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
}

// getAuthors takes a list of commits (c) and returns a slice of those authors.
func getAuthors(c []source.Commit) authorWrappers {
	authors := map[string]authorWrapper{}
	for _, commit := range c {
		if v, ok := authors[commit.Author.Email]; ok {
			authors[commit.Author.Email] = authorWrapper{v.commitCount + 1, commit.Author}
		} else {
			authors[commit.Author.Email] = authorWrapper{1, commit.Author}
		}
	}

	authorList := []authorWrapper{}
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

// getCommits is a healper function that returns all the commits for a
// repostiory, passed as url.
func getCommitsForTag(url string, tagName string) ([]source.Commit, error) {
	repo, err := source.NewInMemRepo(url)
	if err != nil {
		return nil, err
	}

	gm := source.NewGitManager()
	commits, err := gm.GetCommitsForTag(tagName, *repo)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

func (a authorWrappers) Len() int           { return len(a) }
func (a authorWrappers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a authorWrappers) Less(i, j int) bool { return a[j].commitCount < a[i].commitCount }
