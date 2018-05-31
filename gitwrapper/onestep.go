package gitwrapper

import (
	"fmt"
	"io"
	"log"

	"github.com/alecthomas/template"
)

// GithubClone creates a new Repo by cloning from github.
func GithubClone(owner, repo string) (*Repo, error) {
	url := fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	return cloneRepo(url)
}

func (r *Repo) updateVersionFile(newVersion string) error {
	commitMsg := fmt.Sprintf("Change version to %v", newVersion)
	return r.updateFile("version.go", commitMsg, func(w io.Writer) error {
		t := template.Must(template.New("version").Parse(versionTemplate))
		return t.Execute(w, map[string]string{"version": newVersion})
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
	if err := r.push("menghanl", "TODO: pass auth token in"); err != nil {
		log.Fatalf("failed to push: %v", err)
	}
}
