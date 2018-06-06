package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/gitwrapper"
	"golang.org/x/oauth2"
	"gopkg.in/AlecAivazis/survey.v1"

	log "github.com/sirupsen/logrus"
)

var (
	token = flag.String("token", "", "github token")
)

// TODO: make those flags.
const (
	newVersion = "1.new.0"
	username   = "menghanl"
	password   = "TODO: pass auth token in"
)

func main() {
	flag.Parse()

	var tc *http.Client
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	c := ghclient.New(tc, "grpc", "grpc-go")
	// prs := c.GetMergedPRsForMilestone("1.13 Release")
	prs := c.GetMergedPRsForLabels([]string{"Cherry Pick"})
	for i, pr := range prs {
		fmt.Println(i, pr.GetNumber(), pr.GetTitle())
	}
}

func main2() {
	r, err := gitwrapper.GithubClone(&gitwrapper.GithubCloneConfig{
		Owner: "menghanl",
		Repo:  "grpc-go",
	})
	if err != nil {
		log.Fatalf("failed to github clone: %v", err)
	}

	if err := r.MakeVersionChange(&gitwrapper.VersionChangeConfig{
		VersionFile: "version.go",
		NewVersion:  newVersion,
	}); err != nil {
		log.Fatalf("failed to make change: %v", err)
	}

	if err := r.Publish(&gitwrapper.PublicConfig{
		RemoteName: "", // FIXME: specify parent or the fork.
		Auth: &gitwrapper.AuthConfig{
			Username: username,
			Password: password,
		},
	}); err != nil {
		log.Fatalf("failed to public change: %v", err)
	}

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
			Default: "1.12", // TODO: find the next release.
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

	_ = tc

	// TODO: get fork parent.
}
