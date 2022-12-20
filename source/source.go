// source is a package that can retrieve and analyze details around source
// code. Many of the items found here are wrappers on git.
package source

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	CacheDirName     = "proctor"
	CacheRepoDirName = "repos"
)

// ResolveRepoOpts provides instructions for how a repository should be retrieved.
type ResolveRepoOpts struct {
	// instructs doing all retrieval in memory. Note that for medium to large
	// size repos, this can cause significant memory consumption.
	InMemory bool
}

// Tag represents a git tag.
type Tag struct {
	Name string
	Date time.Time
	// the branch a tag is associated with
	Branch string
	// the last, or latest, commit on the tag.
	LastCommit        Hash
	AssociatedCommits []Commit
}

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
			Date: obj.Committer.When,
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

// GetCommitsForTag takes a tagName and its associated repository and returns a
// slice of every commit associated with it. It looks up the commits
// **exclusively** by looking up the Tag.LastCommit field. The commits are
// found by doing the equivelant of a git log against the latest (newest)
// commit associated with the tag. Commits within the slice are arranged in
// order by date. When commits are unable to be retrieved from the repository,
// an error is returned.
func (gm *GitManager) GetCommitsForTag(tagName string, r Repository, opts ...GetCommitsOpts) ([]Commit, error) {

	tags, err := gm.GetTagsFromRepository(r)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving tags for repo: %s. Error: %s", r.URL, err)
	}

	mTags := NewMapOfTags(tags)
	if _, ok := mTags[tagName]; !ok {
		return nil, fmt.Errorf("requsted tag (%s) not found in repo (%s)", tagName, r.URL)
	}
	tag := mTags[tagName]

	emptyHash := Hash{}
	// if r is passed without a ref existent, error immediatly or else a panic
	// (nil pointer) will occur.
	if r.RepoRef == nil {
		return nil, fmt.Errorf("failed to find reference to valid repo when looking up commits.")
	}
	// LastCommit hash was empty and error should be returnd
	if tag.LastCommit == emptyHash {
		return nil, fmt.Errorf("no lastcommit hash was specified with tag.")
	}

	commits, err := r.RepoRef.Log(&git.LogOptions{
		From:  plumbing.Hash(tag.LastCommit),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commits from tag \"%s\". Error from go-git was: %s", tag.Name, err)
	}
	CollectedCommits := []Commit{}
	commits.ForEach(func(o *object.Commit) error {
		CollectedCommits = append(CollectedCommits, Commit{
			Hash: Hash(o.Hash),
			Date: o.Committer.When,
			Author: Person{
				Name:  o.Author.Name,
				Email: o.Author.Email,
			},
			Message: []byte(o.Message),
		})
		return nil
	})

	return CollectedCommits, nil
}

// GetTagsFromRepository accepts a repository returns all tags that are
// associated in it.
func (gm *GitManager) GetTagsFromRepository(r Repository) ([]Tag, error) {
	if r.RepoRef == nil {
		return nil, fmt.Errorf("request to retrieve tags was requested but their was no repo associated with the passed argument")
	}
	tags, err := r.RepoRef.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to retieve tags for repository %s. Error from go-get: %s", r.URL, err)
	}
	var CollectedTags []Tag
	tags.ForEach(func(o *plumbing.Reference) error {

		revision := plumbing.Revision(o.Name().String())
		tagCommitHash, err := r.RepoRef.ResolveRevision(revision)
		// TODO(joshrosso):
		if err != nil {
			fmt.Printf("%s: %s\n", o.Hash(), err)
			return nil
		}
		commitRef, err := r.RepoRef.CommitObject(*tagCommitHash)
		// TODO(joshrosso):
		if err != nil {
			fmt.Println("nah-here")
			return nil
		}

		CollectedTags = append(CollectedTags, Tag{
			Name:       o.Name().Short(),
			LastCommit: Hash(commitRef.Hash),
		})
		return nil
	})

	return CollectedTags, nil
}

// NewMapOfTags returns a map representation of a list of tags where the key is
// set to the tag name.
func NewMapOfTags(t []Tag) map[string]Tag {
	tagsMapped := make(map[string]Tag)
	for _, v := range t {
		tagsMapped[v.Name] = v
	}
	return tagsMapped
}

// ResolveRepo accepts a repository's URL and opts for how the repo should be
// retrieved. By default, it looks up the [getDefaultCacheLocation] to
// determine if the repository was previously cached on the filesystem. If it
// is, it will do a git fetch to grab any new changes and return a reference to
// the repository. If the repo does not exist on the filesystem (cache), it
// will perform a clone that persists it to [getDefaultCacheLocation]. The
// directory name within the cache will be a base64 encoded representation of
// the url.
//
// If you wish to get a repository reference for a repo held entirely in
// memeory, you can set InMemory to true within the [ResolveRepoOpts] argument.
// Note that doing an in-memory clone can consume substatial system resouces
// (heap space) when the repository is large.
func ResolveRepo(url string, opts ...ResolveRepoOpts) (*Repository, error) {
	conf := ResolveRepoOpts{}
	if len(opts) > 0 {
		conf = opts[len(opts)-1]
	}
	if conf.InMemory {
		return newInMemRepo(url)
	}
	// Check for existence of repo in filesystem, if it doesn't exist, clone it;
	// if it does, open and return a ref.
	fp := filepath.Join(getDefaultCacheLocation(), getEncodedCacheName(url))
	if _, err := os.Stat(fp); err != nil {
		fmt.Println("caching repo for the first time, this operation may take a while...")
		return newFSRepo(url)
	}

	ref, err := git.PlainOpen(fp)
	if err != nil {
		return nil, fmt.Errorf("failed opening repo in cache: %s", err)
	}
	err = ref.Fetch(&git.FetchOptions{
		RemoteURL: url,
	})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("failed checking if repo was up to date: %s", err)
		}
	}
	repo := &Repository{
		URL:     url,
		RepoRef: ref,
	}
	return repo, nil
}

// newFSRepo attempts to clone the repository to the filesystem and return a
// reference. If the repo already exists or there is an issue retrieving it
// over the network, an error is returned.
func newFSRepo(url string) (*Repository, error) {
	err := ensureCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed ensuring cache location exists or creating it: %s", err)
	}
	fp := filepath.Join(getDefaultCacheLocation(), getEncodedCacheName(url))
	ref, err := git.PlainClone(fp, true, &git.CloneOptions{
		URL:        url,
		NoCheckout: true,
	})
	if err != nil {
		return nil, err
	}
	repo := &Repository{
		URL:     url,
		RepoRef: ref,
	}
	return repo, nil
}

// newInMemRepo takes the url of a repository, for example
// github.com/spf13/cobra, and constructs an in-memory representation of the
// git-related data. If there is an issue creating this representation, an
// error is returned.
func newInMemRepo(url string) (*Repository, error) {
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

// ensureCacheDir will verify that proctor's cache dir already exists and if it
// doesn't, create it.
func ensureCacheDir() error {
	cacheFp := getDefaultCacheLocation()
	// if specified cache directory does not exist, create it.
	if _, err := os.Stat(cacheFp); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(cacheFp, 0777)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// getDefaultCacheLocation returns $XDG_DATA_HOME/proctor/repos. This is where
// repositories that are cloned (cached) to the filesystem are stored.
func getDefaultCacheLocation() string {
	return filepath.Join(xdg.DataHome, CacheDirName, CacheRepoDirName)
}

// getEncodedCacheName takes a repo's URL and returns its representation in
// base64 encoding. This is used for creating unique cache directories when
// persisting cloned repos onto the filesystem.
func getEncodedCacheName(url string) string {
	return base64.StdEncoding.EncodeToString([]byte(url))
}
