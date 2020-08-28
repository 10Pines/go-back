package backup

import (
	"go-re/internal/stats"
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
		reporter *stats.Reporter
	}

	// Config contains Backup configuration
	Config struct {
		Timestamp    time.Time
		BasePath     string
		WorkerCount  int
		Bucket       string
		BucketRegion string
	}
)

// New creates a new Backup instance
func New(config Config, reporter *stats.Reporter) Backup {
	u := uploader.New(config.Bucket, config.BucketRegion)
	return Backup{
		config:   config,
		uploader: u,
		reporter: reporter,
	}
}

// Process starts the backup pipeline
func (b Backup) Process(repositories []*repository.Repository) {
	b.cloneAndZip(repositories)
	b.uploadToS3(b.config.BasePath)
	b.reporter.Finished()
}

func (b Backup) uploadToS3(workPath string) {
	b.uploader.Sync(workPath)
}

func (b Backup) cloneAndZip(repositories []*repository.Repository) {
	wg, cloneQueue := makeWorkerPool(b.config, b.reporter)
	for _, repo := range repositories {
		cloneQueue <- repo
	}
	close(cloneQueue)
	wg.Wait()
}

func makeWorkerPool(config Config, reporter *stats.Reporter) (*sync.WaitGroup, chan<- *repository.Repository) {
	wg := &sync.WaitGroup{}
	repositories := make(chan *repository.Repository)
	for i := 0; i < config.WorkerCount; i++ {
		go addWorker(repositories, config, wg, reporter)
	}
	return wg, repositories
}

func addWorker(repositories <-chan *repository.Repository, config Config, wg *sync.WaitGroup, reporter *stats.Reporter) {
	wg.Add(1)
	timestamp := config.Timestamp.Format(time.RFC3339)
	w := worker{
		basePath: config.BasePath,
		backupID: timestamp,
	}
	for repo := range repositories {
		if repo.IsEmpty() {
			log.Printf("Skipping Repo[%s] because it's empty", repo.Name())
			break
		}
		repositoryStats := w.Clone(repo)
		reporter.TrackRepository(repositoryStats)
	}
	wg.Done()
}
