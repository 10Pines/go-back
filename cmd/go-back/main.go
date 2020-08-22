package main

import (
	"log"
	"time"

	"github.com/alexflint/go-arg"

	"go-re/internal/uploader"
)

// Well known values that identifies 10Pines repositories
const (
	GitHubOrganizationName = "10Pines"
	GitLabOrganizationID   = 1152254
)

var args struct {
	WorkerCount  int    `default:"8" help:"Number of cloning workers"`
	Bucket       string `arg:"required" help:"S3 bucket name where the backup is stored"`
	Region       string `arg:"required" help:"S3 bucket region"`
	BackupFolder string `arg:"required" help:"Backup will be locally stored inside this folder"`
}

func main() {
	arg.MustParse(&args)
	now := time.Now().UTC().Format(time.RFC3339)

	gh, gl := buildProviders()
	ghRepositories := gh.AllRepositories(GitHubOrganizationName)
	log.Printf("Found %d repository in GitHub", len(ghRepositories))

	glRepositories := gl.AllRepositories(GitLabOrganizationID)
	log.Printf("Found %d repository in GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	b := buildBackup(now)
	b.Process(allRepositories)
	log.Printf("Cloned %d repository", len(allRepositories))

	bucket := uploader.New(args.Bucket, args.Region)
	bucket.Sync(args.BackupFolder)
}
