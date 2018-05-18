// Package gitwrapper wraps around git command.
package gitwrapper

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/alecthomas/template"
	"github.com/menghanl/mydump"
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

func (r *Repo) headCommit() (*object.Commit, error) {
	ref, err := r.r.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to call Head(): %v", err)
	}

	headCommit, err := r.r.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to find commit for head: %v", err)
	}
	return headCommit, nil
}

func (r *Repo) updateVersionFile(newVersion string) error {
	versionFile, err := r.fs.OpenFile("version.go", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file version.go: %v", err)
	}
	defer versionFile.Close()

	t := template.Must(template.New("version").Parse(versionTemplate))
	err = t.Execute(versionFile, map[string]string{"version": newVersion})
	if err != nil {
		return fmt.Errorf("failed to execute template to file: %v", err)
	}
	return nil
}

func (r *Repo) currentStatus() (git.Status, error) {
	worktree, err := r.r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %v", err)
	}
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status from worktree: %v", err)
	}
	return status, nil
}

// Try prints head.
func (r *Repo) Try() {
	headCommit, err := r.headCommit()
	if err != nil {
		log.Fatal(err)
	}
	log.Info(headCommit.String())

	//////////////

	if err := r.updateVersionFile("new-version"); err != nil {
		log.Fatal(err)
	}
	status, err := r.currentStatus()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("%v\n", status)
	mydump.Dump(status)

	log.Info("--- reading new file")
	versionFile, err := r.fs.Open("version.go")
	io.Copy(os.Stdout, versionFile)
}
