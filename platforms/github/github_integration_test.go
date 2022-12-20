//go:build integration

package github

import (
	"testing"
)

const (
	badRepo = "k00/0bernetes/kubernetes"
	k8sRepo = "kubernetes/kubernetes"
)

func TestFailWithBadToken(t *testing.T) {
	token := "badToken"
	conf := GHManagerConfig{
		GHToken: token,
	}
	gm := NewGHManager(conf)

	_, err := gm.GetArtifacts(k8sRepo)
	if err == nil {
		t.Log("fail: expected to receive error from using bad token, but did not")
		t.Fail()
	}
}

func TestFailWithInvalidRepo(t *testing.T) {
	gm := NewGHManager()
	_, err := gm.GetArtifacts(badRepo)
	if err == nil {
		t.Log("fail: expected error from using bad repository, but did not")
		t.Fail()
	}
}

func TestGetArtifacts(t *testing.T) {
	gm := NewGHManager()
	repos, err := gm.GetArtifacts(k8sRepo)
	if err != nil {
		t.Logf("fail: error when trying to retrieve release data: %s", err)
		t.Fail()
	}
	if len(repos) < 1 {
		t.Logf("fail: received %d releases, expected to get greater than 1.", len(repos))
	}
}
