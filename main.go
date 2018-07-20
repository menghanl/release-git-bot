package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/blang/semver"
	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/gitwrapper"
	"github.com/olekukonko/tablewriter"
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

	nokidding = flag.Bool("nokidding", false, "if no kidding, do real release")
)

var (
	upstreamUser = "menghanl" // TODO: change this back to "grpc" by default.
)

func main() {
	flag.Parse()

	if *nokidding {
		upstreamUser = "grpc"
	}

	ver, err := semver.Make(*newVersion)
	if err != nil {
		log.Fatalf("invalid version string %q: %v", *newVersion, err)
	}
	log.Info("version is valid: ", ver.String())

	inputTable := tablewriter.NewWriter(os.Stdout)
	inputTable.SetHeader([]string{"input"})
	inputTable.Append([]string{"user", *user})
	inputTable.Append([]string{"repo", *repo})
	inputTable.Append([]string{"version", *newVersion})
	inputTable.Append([]string{"upstreamRepo", upstreamUser + "/" + *repo})
	inputTable.Render()

	lgty := false
	survey.AskOne(&survey.Confirm{Message: "Looks right?"}, &lgty, nil)
	if !lgty {
		fmt.Printf("Existing")
		return
	}

	var transportClient *http.Client
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		transportClient = oauth2.NewClient(ctx, ts)
	}
	upstreamGithub := ghclient.New(transportClient, upstreamUser, *repo)

	fmt.Printf(" - Cloning %v/%v into memory\n\n", *user, *repo)
	forkLocalGit, err := gitwrapper.GithubClone(&gitwrapper.GithubCloneConfig{
		Owner: *user,
		Repo:  *repo,
	})
	if err != nil {
		log.Fatalf("failed to github clone: %v", err)
	}

	fmt.Println()
	/* Step 1: create an upstream release branch if it doesn't exist */
	upstreamReleaseBranchName := fmt.Sprintf("v%v.%v.x", ver.Major, ver.Minor)
	fmt.Printf(" - Step 1: create an upstream release branch %v/%v/%v\n\n", upstreamUser, *repo, upstreamReleaseBranchName)
	upstreamGithub.NewBranchFromHead(upstreamReleaseBranchName)

	fmt.Println()
	/* Step 2: on release branch, change version file to 1.release.0 */
	fmt.Printf(" - Step 2: on release branch, change version to %v\n\n", *newVersion)
	prURL1 := makePR(upstreamGithub, forkLocalGit, *newVersion, upstreamReleaseBranchName)
	// prURL1 := "https://github.com/menghanl/grpc-go/pull/17"
	fmt.Printf("PR %v created, merge before continuing...\n", prURL1)

	/* Wait for the PR to be merged */
	prMergeConfirmed := false
	for !prMergeConfirmed {
		prompt := &survey.Confirm{
			Message: "Merged?",
		}
		survey.AskOne(prompt, &prMergeConfirmed, nil)
	}

	fmt.Println()
	/* Step 3: generate release note and create draft release */
	fmt.Printf(" - Step 3: generate release note and create draft release\n\n")
	// Get and print the markdown release notes.
	markdownNote := releaseNote(upstreamGithub, ver)
	// fmt.Println(markdownNote)

	releaseTitle := fmt.Sprintf("Release %v", *newVersion)
	releaseURL, err := upstreamGithub.NewDraftRelease("v"+*newVersion, upstreamReleaseBranchName, releaseTitle, markdownNote)
	if err != nil {
		log.Fatal("failed to create release: ", err)
	}
	// releaseURL := "https://github.com/menghanl/grpc-go/release/untaged-blahblahblah"
	fmt.Printf("Draft release %v created, publish before continuing\n", releaseURL)

	/* Wait for the release to be published */
	releasePublishConfirmed := false
	for !releasePublishConfirmed {
		prompt := &survey.Confirm{
			Message: "Published?",
		}
		survey.AskOne(prompt, &releasePublishConfirmed, nil)
	}

	fmt.Println()
	/* Step 4: on release branch, change version file to 1.release.1-dev */
	nextMinorRelease := ver
	nextMinorRelease.Patch++ // Increment the pateh version, not the minor version.
	nextMinorReleaseStr := fmt.Sprintf("%v-dev", nextMinorRelease.String())
	fmt.Printf(" - Step 4: on release branch, change version to %v\n\n", nextMinorReleaseStr)
	// prURL2 := "https://github.com/menghanl/grpc-go/pull/18"
	prURL2 := makePR(upstreamGithub, forkLocalGit, nextMinorReleaseStr, upstreamReleaseBranchName)
	fmt.Println("PR to merge: ", prURL2)

	fmt.Println()
	/* Step 5: on master branch, change version file to 1.release+1.0-dev */
	nextMajorRelease := ver
	nextMajorRelease.Minor++ // Increment the minor version, not the major version.
	nextMajorReleaseStr := fmt.Sprintf("%v-dev", nextMajorRelease.String())
	fmt.Printf(" - Step 5: on master branch, change version to %v\n\n", nextMajorReleaseStr)
	// prURL3 := "https://github.com/menghanl/grpc-go/pull/19"
	prURL3 := makePR(upstreamGithub, forkLocalGit, nextMajorReleaseStr, "master")
	fmt.Println("PR to merge: ", prURL3)

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
		UserName:    "release bot", // TODO: change this!!!
		UserEmail:   "bot@grpc.io",
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
