package backup

import (
	"log"
	"sync"
	"time"

	"go-re/internal/repository"
	"go-re/internal/uploader"
)

type (
	// Backup encapsulates all the steps needed to perform a backup
	Backup struct {
		config   Config
		uploader uploader.Uploader
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
func (b Backup) Process(repositories []*repository.Repository) Stats {
	start := time.Now()
	stats := Stats{}
	stats.RepoStats = b.cloneAndZip(repositories)
	b.uploadToS3(b.config.BasePath)
	stats.Time = time.Since(start).Milliseconds()
	return stats
}

func (b Backup) uploadToS3(workPath string) {
	b.uploader.Sync(workPath)
}

func (b Backup) cloneAndZip(repositories []*repository.Repository) []RepoStats {
	wg, cloneQueue, t := makeWorkerPool(b.config)
	for _, repo := range repositories {
		cloneQueue <- repo
	}
	close(cloneQueue)
	wg.Wait()
	stats := t.Stats()
	log.Printf("cloned %d repositores", len(stats))
	return stats
}

func makeWorkerPool(config Config) (*sync.WaitGroup, chan<- *repository.Repository, *tracker) {
	log.Printf("Clone phase is using %d workers", config.WorkerCount)
	wg := &sync.WaitGroup{}
	t := tracker{}
	repositories := make(chan *repository.Repository)
	for i := 0; i < config.WorkerCount; i++ {
		go addWorker(repositories, config, wg, &t)
	}
	return wg, repositories, &t
}

func addWorker(repositories <-chan *repository.Repository, config Config, wg *sync.WaitGroup, t *tracker) {
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
		result := w.Clone(repo)
		t.Add(result)
	}
	wg.Done()
}

type tracker struct {
	mu      sync.Mutex
	results []RepoStats
}

func (t *tracker) Add(result RepoStats) {
	t.mu.Lock()
	t.results = append(t.results, result)
	t.mu.Unlock()
}

func (t *tracker) Stats() []RepoStats {
	return t.results
}
