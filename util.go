package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/menghanl/release-git-bot/ghclient"
	"github.com/menghanl/release-git-bot/notes"

	log "github.com/sirupsen/logrus"
)

func commaStringToSet(s string) map[string]struct{} {
	ret := make(map[string]struct{})
	tmp := strings.Split(s, ",")
	for _, t := range tmp {
		ret[t] = struct{}{}
	}
	return ret
}

func releaseNote(c *ghclient.Client, ver semver.Version) string {
	milestone := fmt.Sprintf("%v.%v Release", ver.Major, ver.Minor)

	var (
		prs          []*github.Issue
		thanksFilter func(pr *github.Issue) bool
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		prs = c.GetMergedPRsForMilestone(milestone)
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

	ns := notes.GenerateNotes(c.Owner(), c.Repo(), "v"+ver.String(), prs, notes.Filters{
		SpecialThanks: thanksFilter,
	})

	log.Infof("generated notes for %v/%v/%v", c.Owner(), c.Repo(), "v"+ver.String())
	return ns.ToMarkdown()
}
