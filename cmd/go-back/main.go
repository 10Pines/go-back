package main

import (
	"log"
	"time"

	"github.com/10Pines/tracker/pkg/tracker"
)

// Well known values that identifies 10Pines repositories
const (
	GitHubOrganizationName = "10Pines"
	GitLabOrganizationID   = 1152254
)

func main() {
	args := mustParseArgs()
	timestamp := time.Now().UTC()

	log.Printf("using %d workers", args.WorkerCount)
	log.Printf("using %s as working directory", args.BackupFolder)
	log.Printf("repositories will be uploaded to %s", args.Bucket)

	env := mustParseEnv()
	gh, gl := buildProviders(env)

	ghRepositories := gh.AllRepositories(GitHubOrganizationName)
	log.Printf("Found %d repository in GitHub", len(ghRepositories))

	glRepositories := gl.AllRepositories(GitLabOrganizationID)
	log.Printf("Found %d repository in GitLab", len(glRepositories))

	allRepositories := append(ghRepositories, glRepositories...)

	t := tracker.New(env.TrackerAPIKey)

	b := buildBackups(args, t, timestamp)
	b.Process(allRepositories)
}
