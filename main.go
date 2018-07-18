package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/blang/semver"
	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/gitwrapper"
	"golang.org/x/oauth2"
	survey "gopkg.in/AlecAivazis/survey.v1"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

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
	upstreamUser = "menghanl" // TODO: change this back!!! "grpc"
)

func main() {
	flag.Parse()

	ver, err := semver.Make(*newVersion)
	if err != nil {
		log.Fatalf("invalid version string %q: %v", *newVersion, err)
	}
	log.Info("version is valid: ", ver.String())

	var transportClient *http.Client
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		transportClient = oauth2.NewClient(ctx, ts)
	}
	upstreamGithub := ghclient.New(transportClient, upstreamUser, *repo)

	forkLocalGit, err := gitwrapper.GithubClone(&gitwrapper.GithubCloneConfig{
		Owner: *user,
		Repo:  *repo,
	})
	if err != nil {
		log.Fatalf("failed to github clone: %v", err)
	}

	// TODO: more logging to show progress.

	/* Step 1: create an upstream release branch if it doesn't exist */
	upstreamReleaseBranchName := fmt.Sprintf("v%v.%v.x", ver.Major, ver.Minor)
	upstreamGithub.NewBranchFromHead(upstreamReleaseBranchName)

	/* Step 2: send PR to release branch to change version file to 1.release.0 */
	prURL1 := makePR(upstreamGithub, forkLocalGit, *newVersion, upstreamReleaseBranchName)

	prMergeConfirmed := false
	for !prMergeConfirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("PR %v created, merge before continuing. Merged?", prURL1),
		}
		survey.AskOne(prompt, &prMergeConfirmed, nil)
		fmt.Println(prMergeConfirmed)
	}

	/* Step x: generate release note and create draft release */
	// // Get and print the markdown release notes.
	// markdownNote := releaseNote(upstreamGithub, ver)
	// fmt.Println()
	// fmt.Println(markdownNote)
}

func makePR(upstream *ghclient.Client, local *gitwrapper.Repo, newVersionStr, upstreamReleaseBranchName string) string {
	/* Step 1: make version change locally and push to fork */
	branchName := fmt.Sprintf("release_version_%v", newVersionStr)
	if err := local.MakeVersionChange(&gitwrapper.VersionChangeConfig{
		VersionFile: "version.go",
		NewVersion:  newVersionStr,
		BranchName:  branchName,
	}); err != nil {
		log.Fatalf("failed to make change: %v", err)
	}

	if err := local.Publish(&gitwrapper.PublicConfig{
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

	/* Step 2: send pull request to upstream/release_branch with the change */
	prTitle := fmt.Sprintf("Change version to %v", newVersionStr)
	prURL, err := upstream.NewPullRequest(*user, branchName, upstreamReleaseBranchName, prTitle, "")
	if err != nil {
		log.Fatalf("failed to create pull request: ", err)
	}
	return prURL
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
