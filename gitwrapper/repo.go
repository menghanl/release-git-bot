// Package gitwrapper wraps around git command.
package gitwrapper

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/alecthomas/template"
	log "github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Repo represends a git repo.
type Repo struct {
	r        *git.Repository
	worktree *git.Worktree

	fs billy.Filesystem
}

// GithubClone creates a new Repo by cloning from github.
func GithubClone(owner, repo string) (*Repo, error) {
	url := fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	log.Infof("executing %q", "git clone "+url)

	fs := memfs.New()
	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, err
	}

	worktree, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %v", err)
	}

	return &Repo{
		r:        r,
		worktree: worktree,
		fs:       fs,
	}, nil
}

func (r *Repo) createAndCheckoutBranch(name string) error {
	ref, err := r.r.Head()
	if err != nil {
		return fmt.Errorf("failed to call Head(): %v", err)
	}
	log.Infof("HEAD at: %v", ref)

	log.Infof("executing %q", "checkout -b"+name)
	if err := r.worktree.Checkout(&git.CheckoutOptions{
		Hash:   ref.Hash(),
		Branch: plumbing.ReferenceName("refs/heads/" + name),
		Create: true,
	}); err != nil {
		return fmt.Errorf("failed to checkout to new branch: %v", err)
	}

	ref, err = r.r.Head()
	if err != nil {
		return fmt.Errorf("failed to call Head(): %v", err)
	}
	log.Infof("HEAD at: %v", ref)

	commit, err := r.r.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("failed to find commit for head: %v", err)
	}
	log.Infof("Commit at HEAD:\n%v", commit)

	return nil
}

func (r *Repo) updateVersionFile(newVersion string) error {
	log.Infof("executing %q", "edit version.go")
	versionFile, err := r.fs.OpenFile("version.go", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file version.go: %v", err)
	}
	t := template.Must(template.New("version").Parse(versionTemplate))
	err = t.Execute(versionFile, map[string]string{"version": newVersion})
	if err != nil {
		versionFile.Close()
		return fmt.Errorf("failed to execute template to file: %v", err)
	}
	versionFile.Close()

	status, err := r.worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status from worktree: %v", err)
	}
	log.Infof("current worktree status (git status):\n%v", status)

	log.Infof("executing %q", "git commit -m 'Change version to %v'")
	commitMsg := fmt.Sprintf("Change version to %v", newVersion)
	if _, err := r.worktree.Commit(commitMsg, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  "release bot",
			Email: "releasebot@grpc.io",
			When:  time.Now(),
		},
	}); err != nil {
		return fmt.Errorf("failed to commit: %v", err)
	}

	return nil
}

func (r *Repo) printDiffInHeadCommit() error {
	log.Infof("executing %q", "git diff HEAD~")
	headRef, err := r.r.Head()
	if err != nil {
		return fmt.Errorf("failed to call Head(): %v", err)
	}
	headCommit, err := r.r.CommitObject(headRef.Hash())
	if err != nil {
		return fmt.Errorf("failed to get head commit: %v", err)
	}
	parentCommit, err := headCommit.Parent(0)
	if err != nil {
		return fmt.Errorf("failed to get parent of head: %v", err)
	}

	// patch, err := parentTree.Patch(headTree)
	patch, err := parentCommit.Patch(headCommit)
	if err != nil {
		return fmt.Errorf("failed to get patch: %v", err)
	}
	log.Info(patch)

	return nil
}

// Try does the work.
func (r *Repo) Try() {
	// git checkout -b release_version
	if err := r.createAndCheckoutBranch("release_version"); err != nil {
		log.Fatal(err)
	}

	// edit file
	// git commit -m 'Change version to %v'
	const newVersion = "1.new.0"
	if err := r.updateVersionFile(newVersion); err != nil {
		log.Fatal(err)
	}

	// git diff HEAD~
	if err := r.printDiffInHeadCommit(); err != nil {
		log.Fatal(err)
	}

	// git push -u
	if err := r.r.Push(&git.PushOptions{
		Progress: os.Stdout,
	}); err != nil {
		log.Fatalf("failed to push: %v", err)
	}
}
