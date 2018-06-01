package gitwrapper

import (
	"fmt"
	"io"
	"log"
)

const (
	branchName = "release_version"
)

// GithubCloneConfig config github clone.
type GithubCloneConfig struct {
	// Owner is the owner's username on github.
	Owner string
	// Repo is the repo name.
	Repo string
}

// GithubClone creates a new Repo by cloning from github.
func GithubClone(c *GithubCloneConfig) (*Repo, error) {
	url := fmt.Sprintf("https://github.com/%v/%v", c.Owner, c.Repo)
	return cloneRepo(url)
}

// Try does the work.
func (r *Repo) Try() {
	const (
		newVersion = "1.new.0"
		username   = "menghanl"
		password   = "TODO: pass auth token in"
	)
	c := &VersionChangeConfig{
		VersionFile: "version.go",
		NewVersion:  newVersion,

		Username: username,
		Password: password,
	}
	if err := r.MakeVersionChange(c); err != nil {
		log.Fatalf("failed to make change: %v", err)
	}
}

// VersionChangeConfig contains the settings to make a version change.
type VersionChangeConfig struct {
	// VersionFile is the filepath of the version file.
	VersionFile string
	// NewVersion is the new version to be changed to.
	NewVersion string

	// Changes won't be pushed to remote if LocalOnly is true.
	LocalOnly bool
	// Username is the auth username.
	Username string
	// Password is the auth password.
	Password string
}

// MakeVersionChange makes the version change in repo.
func (r *Repo) MakeVersionChange(c *VersionChangeConfig) error {
	// git checkout -b release_version
	if err := r.checkoutBranch(branchName); err != nil {
		return err
	}

	if c.NewVersion == "" {
		return fmt.Errorf("config.NewVersion is empty")
	}
	// edit file
	// git commit -m 'Change version to %v'

	if err := r.updateFile(
		c.VersionFile,
		fmt.Sprintf("Change version to %v", c.NewVersion),
		func(w io.Writer) error {
			return versionTemplate.Execute(w, map[string]string{"version": c.NewVersion})
		},
	); err != nil {
		return err
	}

	// git diff HEAD~
	if err := r.printDiffInHeadCommit(); err != nil {
		return err
	}

	// git push -u
	if err := r.push(c.Username, c.Password); err != nil {
		return err
	}
	return nil
}
