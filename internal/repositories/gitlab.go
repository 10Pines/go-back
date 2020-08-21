package repositories

import (
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gl "github.com/xanzy/go-gitlab"
)

func GetGitlabRepos(auths *Auths, organizationId int) []*Repository {
	client, err := gl.NewClient(auths.GitLabToken)
	gitLabAuth := makeGitLabAuth(auths)
	if err != nil {
		log.Fatal(err)
	}

	_, response, err := client.Groups.ListGroupProjects(organizationId, &gl.ListGroupProjectsOptions{
		ListOptions: gl.ListOptions{
			Page:    1,
			PerPage: PageSize,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	gitLabProjects := doFetchGitlabProjects(client, response.CurrentPage, response.TotalPages)
	return fromGitlabProjects(gitLabProjects, gitLabAuth)
}

func doFetchGitlabProjects(client *gl.Client, currentPage int, lastPage int) []*gl.Project {
	projects, response, err := client.Groups.ListGroupProjects(1152254, &gl.ListGroupProjectsOptions{
		ListOptions: gl.ListOptions{
			Page:    currentPage,
			PerPage: PageSize,
		},
	})
	if err != nil {
		log.Fatalf("Fetching page %d/%d failed", currentPage, lastPage)
	}
	if currentPage != response.TotalPages {
		return append(projects, doFetchGitlabProjects(client, response.NextPage, lastPage)...)
	}
	return projects
}

func makeGitLabAuth(auths *Auths) *http.BasicAuth {
	return &http.BasicAuth{
		Username: "oauth2",
		Password: auths.GitLabToken,
	}
}

func fromGitlabProjects(gitLabProjects []*gl.Project, gitLabAuth *http.BasicAuth) []*Repository {
	var foundRepositories []*Repository
	for _, glProject := range gitLabProjects {
		foundRepositories = append(foundRepositories, &Repository{
			Name:  glProject.Name,
			url:   glProject.HTTPURLToRepo,
			empty: false,
			auth:  gitLabAuth,
			host:  "GitLab",
		})
	}
	return foundRepositories
}
