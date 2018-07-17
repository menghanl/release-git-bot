package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/gitwrapper"
	"github.com/menghanl/release-git-bot/notes"
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

	var (
		owner   = "grpc"
		repo    = "grpc-go"
		release = "v1.12.2"
	)
	fmt.Println(owner, repo, release)

	c := ghclient.New(tc, owner, repo)
	prs := c.GetMergedPRsForMilestone("1.14 Release")
	// prs := c.GetMergedPRsForLabels([]string{"Cherry Pick"})
	// for i, pr := range prs {
	// 	fmt.Println(i, pr.GetNumber(), pr.GetTitle())
	// 	fmt.Println(c.CommitIDForMergedPR(pr))
	// }
	ns := notes.GenerateNotes(owner, repo, release, prs, notes.Filters{})
	fmt.Printf("\n================ generated notes for org %q repo %q release %q ================\n\n", ns.Org, ns.Repo, ns.Version)
	for _, section := range ns.Sections {
		fmt.Printf("# %v\n\n", section.Name)
		for _, entry := range section.Entries {
			fmt.Printf(" * %v (#%v)\n", entry.Title, entry.IssueNumber)
			if entry.SpecialThanks {
				fmt.Printf("   - Special Thanks: @%v\n", entry.User.Login)
			}
		}
		fmt.Println()
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
}

func surveyTemp() {
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
