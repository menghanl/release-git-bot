package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	log "github.com/sirupsen/logrus"
)

var (
	token   = flag.String("token", "", "github token")
	release = flag.String("release", "", "major release number, for example 1.12")
	owner   = flag.String("owner", "grpc", "github repo owner")
	repo    = flag.String("repo", "grpc-go", "github repo")
)

type client struct {
	c *github.Client
}

func (c *client) newBranch(ctx context.Context, branchName string) error {
	log.Infof("creating branch: %v", branchName)
	return fmt.Errorf("unimplemented")
}

func main() {
	flag.Parse()

	if *release == "" {
		fmt.Println("invalid release number, usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var tc *http.Client
	if *token != "" {
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}
	c := &client{c: github.NewClient(tc)}
	if err := c.newBranch(context.Background(), "v"+*release+".x"); err != nil {
		log.Fatal(err)
	}
}
