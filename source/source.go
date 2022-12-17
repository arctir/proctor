// source is a package that can retrieve and analyze details around source
// code. Many of the items found here are wrappers on git.
package source

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Hash [20]byte

type Person struct {
	Name  string
	Email string
}

type Commit struct {
	Hash      Hash
	Title     string
	Date      time.Time
	Committer Person
	Author    Person
	Message   []byte
}

// GitManager operates on [git] repositories in order to facilitate the
// metadata around source repositories. Examples of lookups include [commits],
// [tags], and [artifacts]. This may also extend to the platform hosting git
// repostiories such as GitHub. This is where the asset [artifacts] may be
// retrieved from.
//
// [git]: https://en.wikipedia.org/wiki/Git
// [commits]: https://git-scm.com/docs/git-commit
// [tags]: https://git-scm.com/book/en/v2/Git-Basics-Tagging
// [artifacts]: https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases
type GitManager struct {
	GitManagerConfig
}

// GitManagerConfig provides the configuation settings used to create a
// GitManager. The struct should be created and used when calling the
// [NewGitManager] function.
type GitManagerConfig struct {
	// Represents a [presonal access token] provided by GitHub.
	//
	// [personal access token]: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token
	AccessToken string
}

type CommitReader interface {
	GetCommits() ([]Commit, error)
}

type Repository struct {
	URL     string
	RepoRef *git.Repository
}

// GetCommitsOpts enables putting constraints on the commit data you'd like to
// retrieve.
type GetCommitsOpts struct {
}

// NewGitManager returns and instance of a [GitManager] based on the specified
// config. The config argument is optional. If a config is not passed or
// required values are left out, defaults will be set.
//
// The variadic nature of config is only to facilitate optional config
// arguments. Do not pass more than one instance of config into this function.
// If more than one is passed, the last config in the argument's slice will be
// used.
func NewGitManager(config ...GitManagerConfig) GitManager {
	return GitManager{}
}

// GetCommits takes a [Repository], which should be generated using
// [NewInMemRepo], and provides a slice of commits related to the repository.
// If you'd like to retrieve a subset of commits, an optional opts argument can
// be provided.
//
// If there is an issue retrieving the commits from the repository, an error is
// returned.
func (gm *GitManager) GetCommits(r Repository, opts ...GetCommitsOpts) ([]Commit, error) {
	// if r is passed without a ref existent, return an error immediatly to avoid
	// a panic (nil pointer access).
	if r.RepoRef == nil {
		return nil, fmt.Errorf("failed to find reference to valid repo when looking up commits.")
	}
	commits := []Commit{}
	commitObjs, err := r.RepoRef.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, fmt.Errorf("failed getting all commits from repo. Error from git: %s", err)
	}

	// add each commit object to commits for return
	commitObjs.ForEach(func(obj *object.Commit) error {
		commit := Commit{
			Hash: Hash(obj.Hash),
			Committer: Person{
				Name:  obj.Committer.Name,
				Email: obj.Committer.Email,
			},
			Author: Person{
				Name:  obj.Author.Name,
				Email: obj.Author.Email,
			},
			Message: []byte(obj.Message),
		}
		commits = append(commits, commit)
		return nil
	})

	return commits, nil
}

// NewInMemRepo takes the url of a repository, for example
// github.com/spf13/cobra, and constructs an in-memory representation of the
// git-related data. If there is an issue creating this representation, an
// error is returned.
func NewInMemRepo(url string) (*Repository, error) {
	mStore := memory.NewStorage()
	r, err := git.Clone(mStore, nil, &git.CloneOptions{
		URL:        url,
		NoCheckout: true,
	})
	if err != nil {
		return nil, err
	}

	// 2. Create Repository object and add information from in-memory retrieval.
	remotes, err := r.Remotes()
	if err != nil {
		return nil, err
	}

	if len(remotes) < 1 {
		return nil, fmt.Errorf("Failed creating new in-memory repo object. Could not find atleast one valid remote repostiory")
	}
	repo := &Repository{
		URL:     url,
		RepoRef: r,
	}
	return repo, nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}
