package repository

import (
	"context"
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gh "github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

// GitHub is a GitHub wrapper
type GitHub struct {
	auth  http.BasicAuth
	token string
}

// NewGitHub returns a new wrapper
func NewGitHub(token string) GitHub {
	return GitHub{
		auth:  buildAuthFromToken(token),
		token: token,
	}
}

func buildAuthFromToken(token string) http.BasicAuth {
	return http.BasicAuth{
		Username: token,
		Password: "",
	}
}

// AllRepositories returns all organization's GitHub repositories
func (r GitHub) AllRepositories(organization string) []*Repository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: r.token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := gh.NewClient(tc)

	pageSize := 50
	_, response, err := client.Repositories.ListByOrg(ctx, organization, &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: pageSize,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	gitHubRepositories := doFetchGitHubRepos(ctx, client, response.FirstPage, response.LastPage)
	return r.fromGitHubRepositories(gitHubRepositories)
}

func doFetchGitHubRepos(ctx context.Context, client *gh.Client, currentPage int, lastPage int) []*gh.Repository {
	gitHubRepositories, response, err := client.Repositories.ListByOrg(ctx, "10Pines", &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: pageSize,
			Page:    currentPage,
		},
	})
	if err != nil {
		log.Fatalf("Fetching page %d/%d failed", currentPage, lastPage)
	}
	if currentPage != lastPage {
		return append(gitHubRepositories, doFetchGitHubRepos(ctx, client, response.NextPage, lastPage)...)
	}
	return gitHubRepositories
}

func (r GitHub) fromGitHubRepositories(gitHubRepositories []*gh.Repository) []*Repository {
	var foundRepositories []*Repository
	for _, ghRepo := range gitHubRepositories {
		foundRepositories = append(foundRepositories, &Repository{
			name:  ghRepo.GetName(),
			url:   ghRepo.GetCloneURL(),
			empty: isEmpty(ghRepo),
			auth:  &r.auth,
			host:  "GitHub",
		})
	}
	return foundRepositories
}

func isEmpty(repository *gh.Repository) bool {
	return repository.GetSize() == 0
}
