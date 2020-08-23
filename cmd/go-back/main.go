package main

import (
	"log"
	"os/exec"
	"time"

	"github.com/alexflint/go-arg"
)

// Well known values that identifies 10Pines repositories
const (
	GitHubOrganizationName = "10Pines"
	GitLabOrganizationID   = 1152254
)

type appArgs struct {
	WorkerCount  int    `default:"8" help:"Number of cloning workers"`
	Bucket       string `arg:"required" help:"S3 bucket name where the backup is stored"`
	Region       string `arg:"required" help:"S3 bucket region"`
	BackupFolder string `arg:"required" help:"Backup will be locally stored inside this folder"`
}

func main() {

	cmd := exec.Command("df", "-h")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf(err.Error())
	}
	log.Print(string(output))

	args := appArgs{}
	arg.MustParse(&args)
	now := time.Now().UTC().Format(time.RFC3339)

	log.Printf("using %d workers", args.WorkerCount)
	log.Printf("using %s as working directory", args.BackupFolder)
	log.Printf("repositories will be uploaded to %s", args.Bucket)

	gh, gl := buildProviders()
	ghRepositories := gh.AllRepositories(GitHubOrganizationName)
	log.Printf("Found %d repository in GitHub", len(ghRepositories))

	glRepositories := gl.AllRepositories(GitLabOrganizationID)
	log.Printf("Found %d repository in GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	b := buildBackup(args, now)
	s := b.Process(allRepositories)
	log.Print(s)
}
