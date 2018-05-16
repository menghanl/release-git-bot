// Package gitwrapper wraps around git command.
package gitwrapper

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Repo represends a git repo.
type Repo struct {
	r *git.Repository
}

// GithubClone creates a new Repo by cloning from github.
func GithubClone(owner, repo string) (*Repo, error) {
	url := fmt.Sprintf("https://github.com/%v/%v", owner, repo)

	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		return nil, err
	}

	return &Repo{
		r: r,
	}, nil
}

// PrintHead prints head.
func (r *Repo) PrintHead() {
	// ... retrieves the branch pointed by HEAD
	ref, err := r.r.Head()
	if err != nil {
		log.Fatalf("failed to call Head(): %v", err)
	}

	// ... retrieves the commit history
	cIter, err := r.r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Fatalf("failed to call Log(): %v", err)
	}

	if c, err := cIter.Next(); err == nil {
		log.Info(c.String())
	}
}
