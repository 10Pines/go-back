package main

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gh "github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

const OrganizationName = "10Pines"

func main() {
	ghToken, found := os.LookupEnv("GH_TOKEN")
	if !found {
		log.Fatal("GH_TOKEN is missing")
	}
	auth := makeGithubAuth(ghToken)
	wg := sync.WaitGroup{}
	repositories := make(chan *gh.Repository)
	for i := 0; i < 30; i++ {
		go cloneWorker(repositories, auth, &wg)
	}
	foundRepositories := getGithubRepos(ghToken, OrganizationName)
	for _, repository := range foundRepositories {
		repositories <- repository
	}
	close(repositories)
	wg.Wait()
	log.Printf("Cloned %d repositories", len(foundRepositories))
}

func makeGithubAuth(token string) *http.BasicAuth {
	return &http.BasicAuth{
		Username: token,
		Password: "",
	}
}

func IsEmpty(repository *gh.Repository) bool {
	return repository.GetSize() == 0
}

func cloneWorker(repositories chan *gh.Repository, auth *http.BasicAuth, wg *sync.WaitGroup) {
	wg.Add(1)
	for repo := range repositories {
		if IsEmpty(repo) {
			log.Printf("Skipping Repo[%s] because it's empty", repo.GetName())
			break
		}
		log.Printf("Cloning %s", repo.GetName())
		start := time.Now()
		err := cloneRepo(repo, auth)
		if err != nil {
			log.Fatalf("Failed cloning Repo[%s]. Err[%s]", repo.GetName(), err)
		}
		end := time.Now()
		log.Printf("Cloned %s in %d ms", repo.GetName(), end.Sub(start).Milliseconds())
	}
	wg.Done()
}

func cloneRepo(repository *gh.Repository, auth *http.BasicAuth) error {
	_, err := git.PlainClone(path.Join(".", "repos", repository.GetName()), false, &git.CloneOptions{
		URL:  repository.GetCloneURL(),
		Auth: auth,
	})
	return err
}

func getGithubRepos(token string, organizationName string) []*gh.Repository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
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
	log.Printf("Fetching %d pages @ %d per page", response.LastPage, pageSize)
	return doFetch(client, ctx, response.FirstPage, response.LastPage)
}

func doFetch(client *gh.Client, ctx context.Context, currentPage int, lastPage int) []*gh.Repository {
	foundRepos, response, err := client.Repositories.ListByOrg(ctx, "10Pines", &gh.RepositoryListByOrgOptions{
		ListOptions: gh.ListOptions{
			PerPage: 50,
			Page:    currentPage,
		},
	})
	if err != nil {
		log.Fatalf("Fetching page %d/%d failed", currentPage, lastPage)
	}
	if currentPage != lastPage {
		return append(foundRepos, doFetch(client, ctx, response.NextPage, lastPage)...)
	}
	return foundRepos
}
