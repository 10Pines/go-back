package internal

import (
	"log"
	"os"
	"strconv"
)

const PageSize = 50

type Configuration struct {
	gitLabToken string
	gitHubToken string
	workerCount int
}

func BuildConfiguration() *Configuration {
	gitHubToken, found := os.LookupEnv("GH_TOKEN")
	if !found {
		log.Fatal("GH_TOKEN is missing")
	}
	gitLabToken, found := os.LookupEnv("GL_TOKEN")
	if !found {
		log.Fatal("GL_TOKEN is missing")
	}
	workers, found := os.LookupEnv("WORKERS")
	if !found {
		workers = "10"
	}
	workerCount, err := strconv.Atoi(workers)
	if err != nil {
		log.Fatalf("WORKERS is not a number: %s", workers)
	}

	return &Configuration{
		gitHubToken: gitHubToken,
		gitLabToken: gitLabToken,
		workerCount: workerCount,
	}
}
