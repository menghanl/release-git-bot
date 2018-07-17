package ghclient

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

// Client is a github client used to get info from github.
type Client struct {
	owner string
	repo  string

	c *github.Client
}

// New creates a new client.
func New(tc *http.Client, owner, repo string) *Client {
	return &Client{
		owner: owner,
		repo:  repo,
		c:     github.NewClient(tc),
	}
}

// GetMergedPRsForMilestone returns a list of github issues that are merged PRs
// for this milestone.
func (c *Client) GetMergedPRsForMilestone(milestone string) []*github.Issue {
	return c.getMergedPRsForMilestone(milestone)
}

// GetMergedPRsForLabels returns a list of github issues that are merged PRs
// with the given label.
func (c *Client) GetMergedPRsForLabels(labels []string) []*github.Issue {
	return c.getMergedPRsForLabels(labels)
}

// GetOrgMembers returns a set of names of members in the org.
func (c *Client) GetOrgMembers(org string) map[string]struct{} {
	return c.getOrgMembers(org)
}

// CommitIDForMergedPR returns the commit id for pr.
//
// It returns "" if pr is not a merged PR.
func (c *Client) CommitIDForMergedPR(pr *github.Issue) string {
	return c.commitIDForMergedPR(pr)
}

// NewBranchFromHead create a new branch with the current commit from head. Note
// that owner and repo can be different from Client.
//
// It does nothing if the branch already exists.
func (c *Client) NewBranchFromHead(ctx context.Context, owner, repo, branchName string) error {
	log.Infof("creating branch: %v/%v/%v", owner, repo, branchName)

	refName := "heads/" + branchName
	// Check if ref already exists.
	if ref, _, err := c.c.Git.GetRef(ctx, owner, repo, refName); err == nil {
		log.Infof("ref already exists: %v", ref)
		return nil
	}

	// Get head SHA.
	ref, _, err := c.c.Git.GetRef(ctx, owner, repo, "heads/master")
	if err != nil {
		return fmt.Errorf("failed to get master hash: %v", err)
	}
	log.Infof("hash for HEAD: %v", ref.GetObject().GetSHA())

	// Create new ref.
	newRef, _, err := c.c.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    &refName,
		Object: ref.GetObject(),
	})
	if err != nil {
		return fmt.Errorf("failed to create ref: %v", err)
	}

	log.Infof("new ref created: %v", newRef.String())
	return nil
}
