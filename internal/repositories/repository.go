package repositories

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"go-re/internal/compression"
)

const PageSize = 50

type Repository struct {
	Name  string
	url   string
	empty bool
	auth  *http.BasicAuth
	host  string
}

type CloneConfig struct {
	timestamp   string
	baseFolder  string
	workerCount int
}

func (c *CloneConfig) Path(repository *Repository) string {
	return path.Join(c.baseFolder, repository.host, c.timestamp, repository.Name)
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
	for repository := range repositories {
		if repository.empty {
			log.Printf("Skipping Repo[%s] because it's empty", repository.Name)
			break
		}
		log.Printf("Cloning %s@%s", repository.Name, repository.host)
		start := time.Now()
		repositoryPath := cloneConfig.Path(repository)
		err := cloneRepo(repository, repositoryPath)
		if err != nil {
			log.Fatalf("Failed cloning Repo[%s]. Err[%s]", repository.Name, err)
		}
		end := time.Now()
		log.Printf("Cloned %s in %d ms", repository.Name, end.Sub(start).Milliseconds())
		err = zip(repositoryPath, repository)
		if err != nil {
			log.Fatalf("Failed compressing Repo[%s]. Err[%s]", repository.Name, err)
		}
		err = os.RemoveAll(repositoryPath)
		if err != nil {
			log.Fatalf("Failed deleting Repo[%s]. Err[%s]", repository.Name, err)
		}
	}
	wg.Done()
}

func zip(repositoryPath string, repository *Repository) error {
	repositoriesPath := filepath.Dir(repositoryPath)
	return compression.ZipFolder(repositoryPath, path.Join(repositoriesPath, repository.Name+".zip"))
}

func progress(ctx context.Context, name string) {
	start := time.Now()
	var now time.Time
	for {
		select {
		case <-time.NewTicker(20 * time.Second).C:
			now = time.Now()
			log.Printf("[%s] is still ongoing: %.0fs and ticking", name, now.Sub(start).Seconds())
		case <-ctx.Done():
			return
		}
	}
}

func cloneRepo(repository *Repository, repositoryPath string) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	repositoryName := repository.Name
	go progress(ctx, repositoryName)
	_, err := git.PlainClone(repositoryPath, false, &git.CloneOptions{
		URL:  repository.url,
		Auth: repository.auth,
	})
	cancelCtx()
	return err
}
