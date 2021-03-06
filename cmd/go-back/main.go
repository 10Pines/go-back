package main

import (
	"log"
	"time"

	"github.com/alexflint/go-arg"
)

// Well known values that identifies 10Pines repositories
const (
	GitHubOrganizationName = "10Pines"
	GitLabOrganizationID   = 1152254
)

type appArgs struct {
	WorkerCount      int    `default:"8" help:"Number of cloning workers"`
	Bucket           string `arg:"required" help:"S3 bucket name where the backup is stored"`
	Region           string `arg:"required" help:"S3 bucket region"`
	BackupFolder     string `arg:"required" help:"Backup will be locally stored inside this folder"`
	MetricsNamespace string `arg:"required" help:"Cloudwatch namespace where metrics will be published"`
	MetricsRegion    string `arg:"required" help:"Cloudwatch region where metrics will be published"`
}

func main() {
	args := appArgs{}
	arg.MustParse(&args)
	timestamp := time.Now().UTC()

	log.Printf("using %d workers", args.WorkerCount)
	log.Printf("using %s as working directory", args.BackupFolder)
	log.Printf("repositories will be uploaded to %s", args.Bucket)

	gh, gl := buildProviders()
	ghRepositories := gh.AllRepositories(GitHubOrganizationName)
	log.Printf("Found %d repository in GitHub", len(ghRepositories))

	glRepositories := gl.AllRepositories(GitLabOrganizationID)
	log.Printf("Found %d repository in GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	b := buildBackup(args, timestamp)
	b.Process(allRepositories)
}
