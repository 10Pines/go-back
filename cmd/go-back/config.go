package main

import (
	"go-re/internal/stats"
	"log"
	"os"
	"time"

	"go-re/internal/backup"
	"go-re/internal/repository"
)

type env struct {
	GitLabToken string
	GitHubToken string
}

func buildProviders() (repository.GitHub, repository.GitLab) {
	env := mustParseEnv()
	gh := repository.NewGitHub(env.GitHubToken)
	gl := repository.NewGitLab(env.GitLabToken)
	return gh, gl
}

func buildBackup(args appArgs, timestamp time.Time) backup.Backup {
	config := backup.Config{
		Timestamp:    timestamp,
		BasePath:     args.BackupFolder,
		WorkerCount:  args.WorkerCount,
		Bucket:       args.Bucket,
		BucketRegion: args.Region,
	}
	reporter := stats.NewReporter(config.Timestamp, args.Namespace)
	return backup.New(config, reporter)
}

func mustParseEnv() env {
	gitHubToken, found := os.LookupEnv("GH_TOKEN")
	if !found {
		log.Fatal("GH_TOKEN is missing")
	}
	gitLabToken, found := os.LookupEnv("GL_TOKEN")
	if !found {
		log.Fatal("GL_TOKEN is missing")
	}

	return env{
		GitHubToken: gitHubToken,
		GitLabToken: gitLabToken,
	}
}
