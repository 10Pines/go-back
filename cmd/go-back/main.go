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
	"strconv"
	"sync"
	"time"
)

type Configuration struct {
	gitHubToken string
	workerCount int
}

const OrganizationName = "10Pines"

func main() {
	config := buildConfiguration()
	wg, cloneQueue := buildWorkerPool(config)
	foundRepositories := getGithubRepos(config, OrganizationName)
	for _, repository := range foundRepositories {
		cloneQueue <- repository
	}
	close(cloneQueue)
	wg.Wait()
	log.Printf("Cloned %d cloneQueue", len(foundRepositories))
}

func buildWorkerPool(config *Configuration) (*sync.WaitGroup, chan<- *gh.Repository) {
	auth := makeGithubAuth(config.gitHubToken)
	wg := &sync.WaitGroup{}
	repositories := make(chan *gh.Repository)
	for i := 0; i < config.workerCount; i++ {
		go cloneWorker(repositories, auth, wg)
	}
	return wg, repositories
}

func buildConfiguration() *Configuration {
	token, found := os.LookupEnv("GH_TOKEN")
	if !found {
		log.Fatal("GH_TOKEN is missing")
	}
	workers, found := os.LookupEnv("WORKERS")
	if !found {
		workers = "10"
	}
	workerCount, err := strconv.Atoi(workers)
	if err != nil {
		log.Fatalf("WORKERS is not a number: %s", workers)
	}

	return &Configuration{
		gitHubToken: token,
		workerCount: workerCount,
	}
}

func makeGithubAuth(token string) *http.BasicAuth {
	return &http.BasicAuth{
		Username: token,
		Password: "",
	}
}

func isEmpty(repository *gh.Repository) bool {
	return repository.GetSize() == 0
}

func cloneWorker(repositories <-chan *gh.Repository, auth *http.BasicAuth, wg *sync.WaitGroup) {
	wg.Add(1)
	for repo := range repositories {
		if isEmpty(repo) {
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

func progress(ctx context.Context, name string) {
	start := time.Now()
	var now time.Time
	for {
		select {
		case <-time.Tick(10 * time.Second):
			now = time.Now()
			log.Printf("[%s] is still ongoing: %.0fs and ticking", name, now.Sub(start).Seconds())
		case <-ctx.Done():
			return
		}
	}
}

func cloneRepo(repository *gh.Repository, auth *http.BasicAuth) error {
	ctx := context.Background()
	repositoryName := repository.GetName()
	go progress(ctx, repositoryName)
	_, err := git.PlainClone(path.Join(".", "repos", repositoryName), false, &git.CloneOptions{
		URL:  repository.GetCloneURL(),
		Auth: auth,
	})
	ctx.Done()
	return err
}

func getGithubRepos(config *Configuration, organizationName string) []*gh.Repository {
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
