package source

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
)

const (
	DefaultFilePerms = 0777
	HackDir          = "hack"
	TestingDir       = "test"
	TestingRepoDir   = "repos"
	TestDataDir      = "data-dir"
	CommitMsg1       = "inital commit"
)

func TestNewInMemRepo(t *testing.T) {
	// Verify that invalid repository returns an error
	fakeRepoUrl := "not-real-address"
	resolveOpts := ResolveRepoOpts{InMemory: true}
	_, err := ResolveRepo(fakeRepoUrl, resolveOpts)
	if err == nil {
		t.Logf("fail: Expected error to be returned when attempting to get repo: %s. Not error was returned.", fakeRepoUrl)
		t.Fail()
	}

	// Verify valid repository is returned with correct address
	knownGoodRepo := "https://github.com/EbookFoundation/free-programming-books"
	r, err := ResolveRepo(knownGoodRepo, resolveOpts)
	if err != nil {
		t.Logf("fail: received error when attempt to get repo: %s. Error was: %s.", knownGoodRepo, err)
		t.Fail()
	}

	remote, err := r.RepoRef.Remote("origin")
	if err != nil {
		t.Logf("fail: received error attempting to resolve remots. Error was: %s.", err)
		t.Fail()
	}

	if len(remote.Config().URLs) < 1 {
		t.Log("fail: found 0 URLs associated with the repos remotes.")
		t.Fail()
	} else {
		if remote.Config().URLs[0] != knownGoodRepo {
			t.Logf("fail: found remote's repo to be %s but expected %s.", remote.Config().URLs[0], knownGoodRepo)
			t.Fail()
		}
	}
}

func TestGetCommits(t *testing.T) {
	gm := NewGitManager()

	// empty repository should fail
	_, err := gm.GetCommits(Repository{})
	if err == nil {
		t.Log("fail: GetCommits did not return an error when the repo-ref was nil")
		t.Fail()
	}

	// validate commit is correctly returned
	r, err := createTestRepo1()
	defer cleanTestData()
	if err != nil {
		t.Fatalf("fail: error setting up test repo. error was: %s", err)
	}
	commits, err := gm.GetCommits(*r)
	if err != nil {
		t.Fatalf("fail: error retrieving list of commits from repo: %s", err)
	}
	if len(commits) != 1 {
		t.Fatalf("fail: commit lengh was wrong, expected: %d, actual: %d", 1, len(commits))
	}
	if string(commits[0].Message) != CommitMsg1 {
		t.Fatalf("fail: commit message did not match, expected: %s, actual: %s", CommitMsg1, string(commits[0].Message))
	}
}

func createTestRepo1() (*Repository, error) {
	fp, err := createMockRepoDir("repo1")
	if err != nil {
		return nil, err
	}
	// do initial git init and do not include worktree.
	r, err := git.PlainInit(fp, false)
	if err != nil {
		return nil, err
	}

	createFileInPathWithJunkData(fp, "junkFile1")
	wt, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := wt.Status()
	if err != nil {
		return nil, err
	}

	// add each file
	for file := range status {
		_, err = wt.Add(file)
		if err != nil {
			return nil, err
		}

	}

	_, err = wt.Commit(CommitMsg1, &git.CommitOptions{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		URL:     "fake-url",
		RepoRef: r,
	}, nil
}

func createMockRepoDir(name string) (string, error) {
	fp := getTestRepoDir()
	if _, err := os.Stat(fp); err != nil {
		// if the dir was stat'd (it exists) then remove it.
		if err == nil {
			err = os.Remove(fp)
			// return error if unable to remove existing file
			if err != nil {
				return "", fmt.Errorf("failed cleaning existing testing data directory: %s", err)
			}
		}
	}
	fp = filepath.Join(fp, name)

	err := os.MkdirAll(fp, DefaultFilePerms)
	if err != nil {
		return "", fmt.Errorf("failed creating testing data directory: %s", err)
	}

	return fp, nil
}

func getTestRepoDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(cwd), HackDir, TestingDir, TestDataDir, TestingRepoDir)
}

func createFileInPathWithJunkData(path string, fileName string) error {
	junkData := []byte("asd87ufg890easuf09asdufasd90uf")
	fp := filepath.Join(path, fileName)
	err := os.WriteFile(fp, junkData, DefaultFilePerms)
	if err != nil {
		return err
	}
	return nil
}

func cleanTestData() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("fatal error occured; unable to clean-up test resources, which may invalidate future tests: %s", err)
	}
	fp := filepath.Join(filepath.Dir(cwd), HackDir, TestingDir, TestDataDir)
	err = os.RemoveAll(fp)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalf("fatal error occured; unable to clean-up test resources, which may invalidate future tests: %s", err)
		}
	}
}
