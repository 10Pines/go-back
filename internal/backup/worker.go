// Package backup contains the backup process logic
package backup

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"

	"go-re/internal/compression"
	"go-re/internal/repository"
)

type worker struct {
	basePath string
	backupID string
}

func (w worker) Clone(repository *repository.Repository) {
	repositoryPath := path.Join(w.basePath, repository.Host(), w.backupID, repository.Name())
	w.clone(repository, repositoryPath)
	w.zip(repository.Name(), repositoryPath)
}

func (w worker) zip(repositoryName string, repositoryPath string) {
	log.Printf("Compressing Repo[%s]", repositoryName)
	repositoriesPath := filepath.Dir(repositoryPath)
	err := compression.ZipFolder(repositoryPath, path.Join(repositoriesPath, repositoryName+".zip"))
	if err != nil {
		log.Fatalf("Failed compressing Repo[%s]. Err[%s]", repositoryName, err)
	}
	err = os.RemoveAll(repositoryPath)
	if err != nil {
		log.Fatalf("Failed deleting Repo[%s]. Err[%s]", repositoryName, err)
	}
}

func (w worker) clone(repository *repository.Repository, repositoryPath string) {
	log.Printf("Cloning %s@%s", repository.Name(), repository.Host())
	start := time.Now()
	err := cloneRepo(repository, repositoryPath)
	if err != nil {
		log.Fatalf("Failed cloning Repo[%s]. Err[%s]", repository.Name(), err)
	}
	end := time.Now()
	log.Printf("Cloned %s in %d ms", repository.Name(), end.Sub(start).Milliseconds())
}

func logCloningProgress(ctx context.Context, repositoryName string) {
	start := time.Now()
	var now time.Time
	for {
		select {
		case <-time.NewTicker(20 * time.Second).C:
			now = time.Now()
			log.Printf("[%s] is still ongoing: %.0fs and ticking", repositoryName, now.Sub(start).Seconds())
		case <-ctx.Done():
			return
		}
	}
}

func cloneRepo(repository *repository.Repository, repositoryPath string) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	repositoryName := repository.Name()
	go logCloningProgress(ctx, repositoryName)
	_, err := git.PlainClone(repositoryPath, false, &git.CloneOptions{
		URL:  repository.URL(),
		Auth: repository.Auth(),
	})
	cancelCtx()
	return err
}
