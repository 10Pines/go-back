package main

import (
	"log"
	"os"

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

func buildBackup(timestamp string) backup.Backup {
	config := backup.Config{
		BackupProcessID: timestamp,
		BasePath:        args.BackupFolder,
		WorkerCount:     args.WorkerCount,
	}
	return backup.New(config)
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
