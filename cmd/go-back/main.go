package main

import (
	"log"

	"github.com/alexflint/go-arg"

	"go-re/internal/repositories"
	"go-re/internal/s3"
)

const GitHubOrganizationName = "10Pines"
const GitLabOrganizationId = 1152254

var args struct {
	WorkerCount  int    `default:"8" help:"Number of cloning workers"`
	Bucket       string `arg:"required" help:"S3 bucket name to s3 backups to"`
	Region       string `arg:"required" help:"S3 bucket region"`
	BackupFolder string `arg:"required" help:"Backup will be locally stored inside this folder"`
}

func main() {
	arg.MustParse(&args)
	auths := repositories.MakeAuthsFromEnv()
	cloneConfig := repositories.MakeCloneConfig(args.WorkerCount, args.BackupFolder)

	ghRepositories := repositories.GetGithubRepos(auths, GitHubOrganizationName)
	log.Printf("Fetched %d repositories from GitHub", len(ghRepositories))

	glRepositories := repositories.GetGitlabRepos(auths, GitLabOrganizationId)
	log.Printf("Fetched %d repositories from GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	wg, cloneQueue := repositories.MakeCloneWorkerPool(cloneConfig)
	for _, repository := range allRepositories {
		cloneQueue <- repository
	}

	close(cloneQueue)
	wg.Wait()
	log.Printf("Cloned %d repositories", len(allRepositories))

	bucket := s3.New(args.Bucket, args.Region)
	bucket.Sync(args.BackupFolder)
}
