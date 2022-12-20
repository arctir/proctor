package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type Release struct {
	Name      string
	Tag       string
	Artifacts []Artifact
}

type Artifact struct {
	Name        string
	URL         string
	ContentType string
}

type GHRetriever interface {
	GetArtifacts(repoURL string) []Artifact
}

type GHManager struct {
	GHManagerConfig
	client *github.Client
}

// GHManagerConfig provide configuration options for creating a GitHub Manager.
type GHManagerConfig struct {
	// the access token to use when interacting with GitHub. If you plan to
	// access private repositories, this must be set.
	GHToken string
}

// NewGHManager takes an optional configuration (conf) and returns a
// [GHManager]. If required configuration values are not set, defaults are
// used. While conf is variadic, only the last conf argument passed will be
// used.
func NewGHManager(conf ...GHManagerConfig) GHManager {
	opts := GHManagerConfig{}
	if len(conf) > 0 {
		opts = conf[len(conf)-1]
	}
	var httpClient *http.Client

	// if the GHToken was set, create an HTTP client with the oauth2 token;
	// otherwise nil will be passed.
	if opts.GHToken != "" {
		srcToken := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: opts.GHToken},
		)
		httpClient = oauth2.NewClient(context.Background(), srcToken)
	}
	c := github.NewClient(httpClient)

	return GHManager{GHManagerConfig: opts, client: c}
}

func (g *GHManager) GetArtifacts(repoURL string) ([]Release, error) {
	repo := strings.Split(repoURL, "/")
	if len(repo) < 2 {
		return nil, fmt.Errorf("repoURL (%s) was invalid. Repository should be represented with $ORG_NAME/$REPO_NAME. For example, golang's repo would be (golang/go).", repoURL)
	}
	// TODO(joshrosso): this is where we'll introduce pagination when we're ready.
	releases, _, err := g.client.Repositories.ListReleases(context.Background(), repo[0], repo[1], &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed retrieving releases from GitHub for (%s). Error was: %s", repoURL, err)
	}

	r := []Release{}
	// Print the names and URLs of the downloads (artifacts).
	for _, release := range releases {
		a := []Artifact{}
		for _, asset := range release.Assets {
			a = append(a, Artifact{
				Name:        asset.GetName(),
				URL:         asset.GetURL(),
				ContentType: asset.GetContentType(),
			})
		}
		r = append(r, Release{
			Name:      release.GetName(),
			Tag:       release.GetTagName(),
			Artifacts: a,
		})
	}

	return r, nil
}
