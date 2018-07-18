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

	/* Step 2: on release branch, change version file to 1.release.0 */
	prURL1 := makePR(upstreamGithub, forkLocalGit, *newVersion, upstreamReleaseBranchName)

	/* Wait for the PR to be merged */
	prMergeConfirmed := false
	for !prMergeConfirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("PR %v created, merge before continuing. Merged?", prURL1),
		}
		survey.AskOne(prompt, &prMergeConfirmed, nil)
		fmt.Println(prMergeConfirmed)
	}

	/* Step 3: generate release note and create draft release */
	// Get and print the markdown release notes.
	markdownNote := releaseNote(upstreamGithub, ver)
	fmt.Println()
	fmt.Println(markdownNote)

	releaseTitle := fmt.Sprintf("Release %v", *newVersion)
	releaseURL, err := upstreamGithub.NewDraftRelease("v"+*newVersion, upstreamReleaseBranchName, releaseTitle, markdownNote)
	if err != nil {
		log.Fatal("failed to create release: ", err)
	}

	/* Wait for the release to be published */
	releasePublishConfirmed := false
	for !releasePublishConfirmed {
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Draft release %v created, publish before continuing. Published?", releaseURL),
		}
		survey.AskOne(prompt, &releasePublishConfirmed, nil)
		fmt.Println(releasePublishConfirmed)
	}

	/* Step 4: on release branch, change version file to 1.release.1-dev */
	nextMinorRelease := ver
	nextMinorRelease.Patch++ // Increment the pateh version, not the minor version.
	nextMinorReleaseStr := fmt.Sprintf("%v-dev", nextMinorRelease.String())
	prURL2 := makePR(upstreamGithub, forkLocalGit, nextMinorReleaseStr, upstreamReleaseBranchName)
	fmt.Println("merge PR: ", prURL2)

	/* Step 5: on master branch, change version file to 1.release+1.0-dev */
	nextMajorRelease := ver
	nextMajorRelease.Minor++ // Increment the minor version, not the major version.
	nextMajorReleaseStr := fmt.Sprintf("%v-dev", nextMajorRelease.String())
	prURL3 := makePR(upstreamGithub, forkLocalGit, nextMajorReleaseStr, "master")
	fmt.Println("merge PR: ", prURL3)

	/* Step 6: finish steps as in g3doc */
}

// return value is pr URL.
func makePR(upstream *ghclient.Client, local *gitwrapper.Repo, newVersionStr, upstreamBranchName string) string {
	/* Step 1: make version change locally and push to fork */
	branchName := fmt.Sprintf("release_version_%v", newVersionStr)
	if err := local.MakeVersionChange(&gitwrapper.VersionChangeConfig{
		VersionFile: "version.go",
		NewVersion:  newVersionStr,
		BranchName:  branchName,
		SkipCI:      upstreamBranchName != "master", // Not skip if upstreamBranchName is "master"
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
	prURL, err := upstream.NewPullRequest(*user, branchName, upstreamBranchName, prTitle, "")
	if err != nil {
		log.Fatalf("failed to create pull request: ", err)
	}
	return prURL
}
