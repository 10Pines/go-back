package repository

import (
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gl "github.com/xanzy/go-gitlab"
)

// GitLab is a GitLab wrapper
type GitLab struct {
	token string
	auth  http.BasicAuth
}

// NewGitLab returns a new wrapper
func NewGitLab(token string) GitLab {
	return GitLab{
		token: token,
		auth:  buildGitLabAuth(token),
	}
}

func buildGitLabAuth(token string) http.BasicAuth {
	return http.BasicAuth{
		Username: "oauth2",
		Password: token,
	}
}

// AllRepositories returns all organization's GitLab repositories
func (r GitLab) AllRepositories(organizationID int) []*Repository {
	client, err := gl.NewClient(r.token)

	if err != nil {
		log.Fatal(err)
	}

	_, response, err := client.Groups.ListGroupProjects(organizationID, &gl.ListGroupProjectsOptions{
		ListOptions: gl.ListOptions{
			Page:    1,
			PerPage: pageSize,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	gitLabProjects := doFetchGitlabProjects(client, response.CurrentPage, response.TotalPages)
	return r.fromGitlabProjects(gitLabProjects)
}

func (r GitLab) fromGitlabProjects(gitLabProjects []*gl.Project) []*Repository {
	var foundRepositories []*Repository
	for _, glProject := range gitLabProjects {
		foundRepositories = append(foundRepositories, &Repository{
			name:  glProject.Name,
			url:   glProject.HTTPURLToRepo,
			empty: glProject.DefaultBranch == "",
			auth:  &r.auth,
			host:  "GitLab",
		})
	}
	return foundRepositories
}

func doFetchGitlabProjects(client *gl.Client, currentPage int, lastPage int) []*gl.Project {
	projects, response, err := client.Groups.ListGroupProjects(1152254, &gl.ListGroupProjectsOptions{
		ListOptions: gl.ListOptions{
			Page:    currentPage,
			PerPage: pageSize,
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
