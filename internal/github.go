package internal

import (
	"context"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gh "github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"log"
)

func GetGithubRepos(config *Configuration, organizationName string) []*Repository {
	gitHubAuth := makeGithubAuth(config)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.gitHubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := gh.NewClient(tc)

	pageSize := 50
	_, response, err := client.Repositories.ListByOrg(ctx, organizationName, &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: pageSize,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	gitHubRepositories := doFetchGitHubRepos(client, ctx, response.FirstPage, response.LastPage)
	return fromGitHubRepositories(gitHubRepositories, gitHubAuth)
}

func doFetchGitHubRepos(client *gh.Client, ctx context.Context, currentPage int, lastPage int) []*gh.Repository {
	gitHubRepositories, response, err := client.Repositories.ListByOrg(ctx, "10Pines", &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: PageSize,
			Page:    currentPage,
		},
	})
	if err != nil {
		log.Fatalf("Fetching page %d/%d failed", currentPage, lastPage)
	}
	if currentPage != lastPage {
		return append(gitHubRepositories, doFetchGitHubRepos(client, ctx, response.NextPage, lastPage)...)
	}
	return gitHubRepositories
}

func fromGitHubRepositories(gitHubRepositories []*gh.Repository, gitHubAuth *http.BasicAuth) []*Repository {
	var foundRepositories []*Repository
	for _, ghRepo := range gitHubRepositories {
		foundRepositories = append(foundRepositories, &Repository{
			name:  ghRepo.GetName(),
			url:   ghRepo.GetCloneURL(),
			empty: isEmpty(ghRepo),
			auth:  gitHubAuth,
			host: "GitHub",
		})
	}
	return foundRepositories
}

func isEmpty(repository *gh.Repository) bool {
	return repository.GetSize() == 0
}

func makeGithubAuth(configuration *Configuration) *http.BasicAuth {
	return &http.BasicAuth{
		Username: configuration.gitHubToken,
		Password: "",
	}
}
