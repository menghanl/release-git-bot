// Package gitwrapper wraps around git command.
package gitwrapper

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Repo represends a git repo.
type Repo struct {
	r  *git.Repository
	fs billy.Filesystem
}

// GithubClone creates a new Repo by cloning from github.
func GithubClone(owner, repo string) (*Repo, error) {
	url := fmt.Sprintf("https://github.com/%v/%v", owner, repo)

	fs := memfs.New()
	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		return nil, err
	}

	return &Repo{
		r:  r,
		fs: fs,
	}, nil
}

// Try prints head.
func (r *Repo) Try() {
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

	c, err := cIter.Next()
	if err != nil {
		log.Fatalf("failed to get first commit: %v", err)
	}
	log.Info(c.String())

	readme, _ := r.fs.Open("README.md")
	io.Copy(os.Stdout, readme)
}
