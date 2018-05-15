// Package gitwrapper wraps around git command.
package gitwrapper

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Try is called by main to try commands.
func Try() {
	owner := "menghanl"
	repo := "release-note-gen"
	url := fmt.Sprintf("https://github.com/%v/%v", owner, repo)

	// Clones the given repository in memory, creating the remote, the local
	// branches and fetching the objects, exactly as:
	log.Infof("git clone %v", url)

	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL: url,
	})

	if err != nil {
		log.Fatalf("failed to clone: %v", err)
	}

	// Gets the HEAD history from HEAD, just like does:
	log.Info("git log")

	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	if err != nil {
		log.Fatalf("failed to call Head(): %v", err)
	}

	// ... retrieves the commit history
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Fatalf("failed to call Log(): %v", err)
	}

	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
	if err != nil {
		log.Fatalf("failed to call iter.ForEach(): %v", err)
	}
}
