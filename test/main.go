// test is a testing only binary to test release note generating.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"github.com/blang/semver"
	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/notes"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	token = flag.String("token", "", "github token")
)

func main() {
	flag.Parse()
	const (
		upstreamUser = "grpc"
		repo         = "grpc-go"
		version      = "1.39.0"
	)

	ver, err := semver.Make(version)
	if err != nil {
		log.Fatalf("invalid version string %q: %v", version, err)
	}
	log.Info("version is valid: ", ver.String())

	milestone := fmt.Sprintf("%v.%v Release", ver.Major, ver.Minor)

	var transportClient *http.Client
	fmt.Println(*token)
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		transportClient = oauth2.NewClient(ctx, ts)
	}
	c := ghclient.New(transportClient, upstreamUser, repo)

	prs := c.GetMergedPRsForMilestone(milestone)
	ns := notes.GenerateNotes(c.Owner(), c.Repo(), "v"+ver.String(), prs, notes.Filters{})

	fmt.Println(ns.ToMarkdown())
}
