package repositories

import (
	"context"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"log"
	"path"
	"sync"
	"time"
)

const PageSize = 50

type Repository struct {
	Name  string
	url   string
	empty bool
	auth  *http.BasicAuth
	from  string
	host  string
}

type CloneConfig struct {
	timestamp   string
	baseFolder  string
	workerCount int
}

func (c *CloneConfig) ClonePath(repository *Repository) string {
	return path.Join(c.baseFolder, "clone", repository.host, c.timestamp, repository.Name)
}

func (c *CloneConfig) ZipPath(repository *Repository) string {
	return path.Join(c.baseFolder, "zip", repository.host, c.timestamp)
}

func MakeCloneConfig(workerCount int, baseFolder string) *CloneConfig {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	return &CloneConfig{
		timestamp:   timestamp,
		baseFolder:  baseFolder,
		workerCount: workerCount,
	}
}

func MakeCloneWorkerPool(config *CloneConfig) (*sync.WaitGroup, chan<- *Repository) {
	log.Printf("Clone phase is using %d workers", config.workerCount)
	wg := &sync.WaitGroup{}
	repositories := make(chan *Repository)
	for i := 0; i < config.workerCount; i++ {
		go cloneWorker(repositories, config, wg)
	}
	return wg, repositories
}

func cloneWorker(repositories <-chan *Repository, cloneConfig *CloneConfig, wg *sync.WaitGroup) {
	wg.Add(1)
	for repo := range repositories {
		if repo.empty {
			log.Printf("Skipping Repo[%s] because it's empty", repo.Name)
			break
		}
		log.Printf("Cloning %s@%s", repo.Name, repo.host)
		start := time.Now()
		err := cloneRepo(repo, cloneConfig)
		if err != nil {
			log.Fatalf("Failed cloning Repo[%s]. Err[%s]", repo.Name, err)
		}
		end := time.Now()
		log.Printf("Cloned %s in %d ms", repo.Name, end.Sub(start).Milliseconds())
	}
	wg.Done()
}

func progress(ctx context.Context, name string) {
	start := time.Now()
	var now time.Time
	for {
		select {
		case <-time.Tick(20 * time.Second):
			now = time.Now()
			log.Printf("[%s] is still ongoing: %.0fs and ticking", name, now.Sub(start).Seconds())
		case <-ctx.Done():
			return
		}
	}
}

func cloneRepo(repository *Repository, config *CloneConfig) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	repositoryName := repository.Name
	go progress(ctx, repositoryName)
	_, err := git.PlainClone(config.ClonePath(repository), false, &git.CloneOptions{
		URL:  repository.url,
		Auth: repository.auth,
	})
	cancelCtx()
	return err
}
