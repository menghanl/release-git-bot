package gitwrapper

import (
	"fmt"
	"io"
)

// AuthConfig configures auth.
type AuthConfig struct {
	// Username is the auth username.
	Username string
	// Password is the auth password.
	Password string
}

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

// VersionChangeConfig contains the settings to make a version change.
type VersionChangeConfig struct {
	// VersionFile is the filepath of the version file.
	VersionFile string
	// NewVersion is the new version to be changed to. It's a string so it could
	// contain "-dev".
	NewVersion string
	// BranchName is the branch where the change will be made.
	BranchName string
	// SkipCI controls whether travis tests will be skipped.
	SkipCI bool

	// Changes won't be pushed to remote if LocalOnly is true.
	LocalOnly bool
}

// MakeVersionChange makes the version change in repo.
func (r *Repo) MakeVersionChange(c *VersionChangeConfig) error {
	// git checkout master, all changes should be based on master.
	if err := r.checkoutBranch("master"); err != nil {
		return err
	}
	// git checkout -b release_version_1.14.0
	if err := r.checkoutBranch(c.BranchName); err != nil {
		return err
	}

	if c.NewVersion == "" {
		return fmt.Errorf("config.NewVersion is empty")
	}
	// edit file
	// git commit -m 'Change version to %v'
	commitMsg := fmt.Sprintf("Change version to %v", c.NewVersion)
	if c.SkipCI {
		commitMsg += "\n\n[skip ci] Skipping Travis. Version number change only"
	}
	if err := r.updateFile(
		c.VersionFile,
		commitMsg,
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
	return nil
}

// PublicConfig configures public.
type PublicConfig struct {
	// The remote to be pushed to.
	RemoteName string
	// The config for auth.
	Auth *AuthConfig
}

// Publish pushes the local change.
func (r *Repo) Publish(c *PublicConfig) error {
	// This could push to upstream directly, but to be safe, we send pull
	// request instead.

	// git push -u
	if err := r.push(c.Auth.Username, c.Auth.Password); err != nil {
		return err
	}
	return nil
}
