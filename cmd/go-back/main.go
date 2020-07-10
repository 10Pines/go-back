package main

import (
	. "go-re/internal"
	"log"
)

const GitHubOrganizationName = "10Pines"
const GitLabOrganizationId = 1152254

func main() {
	config := BuildConfiguration()
	wg, cloneQueue := BuildWorkerPool(config)
	ghRepositories := GetGithubRepos(config, GitHubOrganizationName)
	log.Printf("Cloning %d repositories from GitHub", len(ghRepositories))
	glRepositories := GetGitlabRepos(config, GitLabOrganizationId)
	log.Printf("Cloning %d repositories from GitLab", len(glRepositories))
	allRepositories := append(ghRepositories, glRepositories...)
	for _, repository := range allRepositories {
		cloneQueue <- repository
	}
	close(cloneQueue)
	wg.Wait()
	log.Printf("Finished %d repos", len(ghRepositories))
}
