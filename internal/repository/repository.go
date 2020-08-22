// Package repository contains common repository logic
package repository

import (
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const pageSize = 50

// Repository is a repository
type Repository struct {
	name  string
	url   string
	empty bool
	auth  *http.BasicAuth
	host  string
}

// Name returns repository name
func (r Repository) Name() string {
	return r.name
}

// Host returns repository host
func (r Repository) Host() string {
	return r.host
}

// URL returns repository URL
func (r Repository) URL() string {
	return r.url
}

// Auth returns repository authentication
func (r Repository) Auth() *http.BasicAuth {
	return r.auth
}

// IsEmpty indicates whether this repository is empty or not
func (r Repository) IsEmpty() bool {
	return r.empty
}
