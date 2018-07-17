package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/gitwrapper"
	"github.com/menghanl/release-git-bot/notes"
	"golang.org/x/oauth2"

	log "github.com/sirupsen/logrus"
)

var (
	token      = flag.String("token", "", "github token")
	newVersion = flag.String("version", "", "the new version number, in the format of Major.Minor.Patch, e.g. 1.14.0")
	user       = flag.String("user", "menghanl", "the github user. Changes will be made this user's fork")
	repo       = flag.String("repo", "grpc-go", "the repo this release is for, e.g. grpc-go")

	// For specials thanks note.
	thanks    = flag.Bool("thanks", true, "whether to include thank you note. grpc organization members are excluded")
	urwelcome = flag.String("urwelcome", "", "list of users to exclude from thank you note, format: user1,user2")
	verymuch  = flag.String("verymuch", "", "list of users to include in thank you note even if they are grpc org members, format: user1,user2")
)

const (
	upstreamUser = "grpc"
)

func commaStringToSet(s string) map[string]struct{} {
	ret := make(map[string]struct{})
	tmp := strings.Split(s, ",")
	for _, t := range tmp {
		ret[t] = struct{}{}
	}
	return ret
}

func main() {
	flag.Parse()

	ver, err := semver.Make(*newVersion)
	if err != nil {
		log.Fatalf("invalid version string %q: %v", *newVersion, err)
	}
	milestoneName := fmt.Sprintf("%v.%v Release", ver.Major, ver.Minor)

	var tc *http.Client
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	c := ghclient.New(tc, upstreamUser, *repo)
	var (
		prs          []*github.Issue
		thanksFilter func(pr *github.Issue) bool
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		prs = c.GetMergedPRsForMilestone(milestoneName)
		wg.Done()
	}()
	if *thanks {
		wg.Add(1)
		go func() {
			urwelcomeMap := commaStringToSet(*urwelcome)
			verymuchMap := commaStringToSet(*verymuch)
			grpcMembers := c.GetOrgMembers("grpc")
			thanksFilter = func(pr *github.Issue) bool {
				user := pr.GetUser().GetLogin()
				_, isGRPCMember := grpcMembers[user]
				_, isWelcome := urwelcomeMap[user]
				_, isVerymuch := verymuchMap[user]
				return *thanks && (isVerymuch || (!isGRPCMember && !isWelcome))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	ns := notes.GenerateNotes(upstreamUser, *repo, "v"+*newVersion, prs, notes.Filters{
		SpecialThanks: thanksFilter,
	})

	fmt.Printf("\n================ generated notes for %v/%v/%v ================\n\n", ns.Org, ns.Repo, ns.Version)
	fmt.Println(ns.ToMarkdown())
}

// Function to make code version changes.
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
		NewVersion:  *newVersion,
	}); err != nil {
		log.Fatalf("failed to make change: %v", err)
	}

	if err := r.Publish(&gitwrapper.PublicConfig{
		// This could push to upstream directly, but to be safe, we send pull
		// request instead.
		RemoteName: "",
		Auth: &gitwrapper.AuthConfig{
			Username: *user,
			Password: *token,
		},
	}); err != nil {
		log.Fatalf("failed to public change: %v", err)
	}

	return
}

// This function is unused.
// func surveyTemp() {
// 	flag.Parse()

// 	qs := []*survey.Question{{
// 		Name: "owner",
// 		Prompt: &survey.Input{
// 			Message: "Who is the owner of the repo?",
// 			Default: "menghanl",
// 		},
// 		Validate: survey.Required,
// 	}, {
// 		Name: "repo",
// 		Prompt: &survey.Input{
// 			Message: "What is the name of the repo?",
// 			Default: "release-note-gen",
// 		},
// 		Validate: survey.Required,
// 	}, {
// 		Name: "release",
// 		Prompt: &survey.Input{
// 			Message: "What is the major release number (e.g. 1.12)?",
// 			Help:    "Only the major release number, without v, without minor release number",
// 			Default: "1.12", // TODO: find the next release.
// 		},
// 		Validate: survey.Required, // TODO: release number validator.
// 	}}

// 	answers := struct {
// 		Owner   string
// 		Repo    string
// 		Release string
// 	}{}

// 	if err := survey.Ask(qs, &answers); err != nil {
// 		log.Fatal(err)
// 		return
// 	}

// 	log.Infof("%v", answers)

// 	var tc *http.Client
// 	if *token != "" {
// 		ctx := context.Background()
// 		ts := oauth2.StaticTokenSource(
// 			&oauth2.Token{AccessToken: *token},
// 		)
// 		tc = oauth2.NewClient(ctx, ts)
// 	}

// 	_ = tc

// 	// TODO: get fork parent.
// }

// TODO: support minor releases
// prs := c.GetMergedPRsForLabels([]string{"Cherry Pick"})
// for i, pr := range prs {
// 	fmt.Println(i, pr.GetNumber(), pr.GetTitle())
// 	fmt.Println(c.CommitIDForMergedPR(pr)) // Get commit ids for cherry-pick.
// }
