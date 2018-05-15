// Package ghwrapper wraps around github APIs.
package ghwrapper

import (
	"context"
	"fmt"

	"github.com/google/go-github/github"
	"github.com/prometheus/common/log"
)

// Client is a github client wrapper.
type Client struct {
	c *github.Client
}

// NewClient creates a new client.
func NewClient(c *github.Client) *Client {
	return &Client{c: c}
}

// NewBranchFromHead create a new branch with the current commit from head.
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
