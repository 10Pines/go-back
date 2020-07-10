package internal

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"log"
	"path"
	"sync"
	"time"
)

type Repository struct {
	name  string
	url   string
	empty bool
	auth  *http.BasicAuth
	from  string
	host  string
}

func BuildWorkerPool(config *Configuration) (*sync.WaitGroup, chan<- *Repository) {
	wg := &sync.WaitGroup{}
	repositories := make(chan *Repository)
	for i := 0; i < config.workerCount; i++ {
		go cloneWorker(repositories, wg)
	}
	return wg, repositories
}

func cloneWorker(repositories <-chan *Repository, wg *sync.WaitGroup) {
	wg.Add(1)
	for repo := range repositories {
		if repo.empty {
			log.Printf("Skipping Repo[%s] because it's empty", repo.name)
			break
		}
		log.Printf("Cloning %s", repo.name)
		start := time.Now()
		err := cloneRepo(repo)
		if err != nil {
			log.Fatalf("Failed cloning Repo[%s]. Err[%s]", repo.name, err)
		}
		end := time.Now()
		log.Printf("Cloned %s in %d ms", repo.name, end.Sub(start).Milliseconds())
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

func cloneRepo(repository *Repository) error {
	ctx := context.Background()
	repositoryName := repository.name
	go progress(ctx, repositoryName)
	_, err := git.PlainClone(path.Join("repos", repository.host, repositoryName), false, &git.CloneOptions{
		URL:  repository.url,
		Auth: repository.auth,
	})
	ctx.Done()
	return err
}
