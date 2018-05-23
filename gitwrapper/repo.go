// Package gitwrapper wraps around git command.
package gitwrapper

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"github.com/alecthomas/template"
	log "github.com/sirupsen/logrus"
	billy "gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/memfs"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
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
	gitdir, err := fs.Chroot(".git")
	if err != nil {
		return nil, fmt.Errorf("failed to chroot(.git): %v", err)
	}
	s, err := filesystem.NewStorage(gitdir)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %v", err)
	}
	r, err := git.Clone(s, fs, &git.CloneOptions{
		URL: url,
		// Only fetch master branch.
		ReferenceName: plumbing.Master,
		SingleBranch:  true,
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

// checkoutBranch checks out to the given branch.
//
// If the branch doesn't exist, a new one will be created.
func (r *Repo) checkoutBranch(name string) error {
	head, err := r.r.Head()
	if err != nil {
		return fmt.Errorf("failed to call Head(): %v", err)
	}
	log.Infof("HEAD at: %v", head)

	newRefName := plumbing.ReferenceName("refs/heads/" + name)
	if _, err := r.r.Reference(newRefName, false); err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return fmt.Errorf("failed to find ref: %v", err)
		}
		// If new ref doesn't exist, create it.
		log.Infof("executing %q", "branch "+name)
		newRef := plumbing.NewHashReference(newRefName, head.Hash())
		if err := r.r.Storer.SetReference(newRef); err != nil {
			return fmt.Errorf("failed to add ref to storer: %v", err)
		}
	}

	log.Infof("executing %q", "checkout "+name)
	if err := r.worktree.Checkout(&git.CheckoutOptions{
		Branch: newRefName,
	}); err != nil {
		return fmt.Errorf("failed to checkout to new branch: %v", err)
	}

	newHead, err := r.r.Head()
	if err != nil {
		return fmt.Errorf("failed to call Head(): %v", err)
	}
	log.Infof("HEAD at: %v", newHead)

	commit, err := r.r.CommitObject(newHead.Hash())
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
	r.worktree.Add("version.go")
	log.Infof("current worktree status (git status):\n%v", status)

	log.Infof("executing %q", "git commit -m 'Change version to %v'")
	commitMsg := fmt.Sprintf("Change version to %v", newVersion)
	if _, err := r.worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			// TODO: change name and email here.
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

// For debugging only.
func (r *Repo) printRepoInfo() {
	refs, _ := r.r.References()
	refs.ForEach(func(r *plumbing.Reference) error {
		fmt.Println(r)
		return nil
	})
}

// Try does the work.
func (r *Repo) Try() {
	// git checkout -b release_version
	if err := r.checkoutBranch("release_version"); err != nil {
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
		Auth: &http.BasicAuth{
			Username: "menghanl",
			Password: "TODO: pass auth token in",
		},
		Progress: os.Stdout,
	}); err != nil {
		log.Fatalf("failed to push: %v", err)
	}
}
