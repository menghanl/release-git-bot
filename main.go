package main

import (
	"context"
	"flag"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/menghanl/release-git-bot/ghwrapper"
	"github.com/menghanl/release-git-bot/gitwrapper"
	"golang.org/x/oauth2"
	"gopkg.in/AlecAivazis/survey.v1"

	log "github.com/sirupsen/logrus"
)

var (
	token = flag.String("token", "", "github token")
)

func main() {
	r, err := gitwrapper.GithubClone("menghanl", "release-note-gen")
	if err != nil {
		log.Fatalf("failed to github clone: %v", err)
	}
	r.Try()
	return

	flag.Parse()

	qs := []*survey.Question{{
		Name: "owner",
		Prompt: &survey.Input{
			Message: "Who is the owner of the repo?",
			Default: "menghanl",
		},
		Validate: survey.Required,
	}, {
		Name: "repo",
		Prompt: &survey.Input{
			Message: "What is the name of the repo?",
			Default: "release-note-gen",
		},
		Validate: survey.Required,
	}, {
		Name: "release",
		Prompt: &survey.Input{
			Message: "What is the major release number (e.g. 1.12)?",
			Help:    "Only the major release number, without v, without minor release number",
			Default: "1.12", // TODO: remove default.
		},
		Validate: survey.Required, // TODO: release number validator.
	}}

	answers := struct {
		Owner   string
		Repo    string
		Release string
	}{}

	if err := survey.Ask(qs, &answers); err != nil {
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

	c := ghwrapper.NewClient(github.NewClient(tc))
	if err := c.NewBranchFromHead(context.Background(), answers.Owner, answers.Repo, "v"+answers.Release+".x"); err != nil {
		log.Fatal(err)
	}
}
