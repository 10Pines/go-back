package repositories

import (
	"log"
	"os"
)

type Auths struct {
	GitLabToken string
	GitHubToken string
}

func MakeAuthsFromEnv() *Auths {
	gitHubToken, found := os.LookupEnv("GH_TOKEN")
	if !found {
		log.Fatal("GH_TOKEN is missing")
	}
	gitLabToken, found := os.LookupEnv("GL_TOKEN")
	if !found {
		log.Fatal("GL_TOKEN is missing")
	}

	return &Auths{
		GitHubToken: gitHubToken,
		GitLabToken: gitLabToken,
	}
}
