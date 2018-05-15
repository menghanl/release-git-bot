package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"gopkg.in/AlecAivazis/survey.v1"

	log "github.com/sirupsen/logrus"
)

type client struct {
	c *github.Client
}

func (c *client) newBranchFromHead(ctx context.Context, owner, repo, branchName string) error {
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

var (
	token = flag.String("token", "", "github token")
)

var qs = []*survey.Question{
	{
		Name: "owner",
		Prompt: &survey.Input{
			Message: "Who is the owner of the repo?",
			Default: "menghanl",
		},
		Validate: survey.Required,
	},
	{
		Name: "repo",
		Prompt: &survey.Input{
			Message: "What is the name of the repo?",
			Default: "release-note-gen",
		},
		Validate: survey.Required,
	},
	{
		Name: "release",
		Prompt: &survey.Input{
			Message: "What is the major release number (e.g. 1.12)?",
			Help:    "Only the major release number, without v, without minor release number",
			Default: "1.12", // TODO: remove default.
		},
		Validate: survey.Required, // TODO: release number validator.
	},
}

func main() {
	flag.Parse()

	answers := struct {
		Owner   string
		Repo    string
		Release string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Infof("%v", answers)

	var tc *http.Client
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	c := &client{c: github.NewClient(tc)}
	if err := c.newBranchFromHead(context.Background(), answers.Owner, answers.Repo, "v"+answers.Release+".x"); err != nil {
		log.Fatal(err)
	}
}
