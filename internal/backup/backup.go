package backup

import (
	"log"
	"sync"

	"go-re/internal/repository"
)

type (
	// Backup encapsulates all the steps needed to perform a backup
	Backup struct {
		config Config
	}

	// Config contains Backup configuration
	Config struct {
		BackupProcessID string
		BasePath        string
		WorkerCount     int
	}
)

// New creates a new Backup instance
func New(config Config) Backup {
	return Backup{
		config: config,
	}
}

// Process starts the backup pipeline
func (b Backup) Process(repositories []*repository.Repository) {
	wg, cloneQueue := makeWorkerPool(b.config)
	for _, repo := range repositories {
		cloneQueue <- repo
	}
	close(cloneQueue)
	wg.Wait()
}

func makeWorkerPool(config Config) (*sync.WaitGroup, chan<- *repository.Repository) {
	log.Printf("Clone phase is using %d workers", config.WorkerCount)
	wg := &sync.WaitGroup{}
	repositories := make(chan *repository.Repository)
	for i := 0; i < config.WorkerCount; i++ {
		go addWorker(repositories, config, wg)
	}
	return wg, repositories
}

func addWorker(repositories <-chan *repository.Repository, config Config, wg *sync.WaitGroup) {
	wg.Add(1)
	w := worker{
		basePath: config.BasePath,
		backupID: config.BackupProcessID,
	}
	for repo := range repositories {
		if repo.IsEmpty() {
			log.Printf("Skipping Repo[%s] because it's empty", repo.Name())
			break
		}
		w.Clone(repo)
	}
	wg.Done()
}
