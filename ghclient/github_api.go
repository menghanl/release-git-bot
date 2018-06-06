package ghclient

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/google/go-github/github"
	log "github.com/sirupsen/logrus"
)

func (c *Client) getMilestoneNumberForTitle(ctx context.Context, milestoneTitle string) (int, error) {
	log.Info("milestone title: ", milestoneTitle)
	milestones, _, err := c.c.Issues.ListMilestones(context.Background(), c.owner, c.repo,
		&github.MilestoneListOptions{
			State:       "all",
			ListOptions: github.ListOptions{PerPage: 100},
		},
	)
	if err != nil {
		return 0, err
	}
	log.Info("count milestones", len(milestones))
	for _, m := range milestones {
		if m.GetTitle() == milestoneTitle {
			return m.GetNumber(), nil
		}
	}
	return 0, fmt.Errorf("no milestone with title %q was found", milestoneTitle)
}

func (c *Client) getMergeEventForPR(ctx context.Context, issue *github.Issue) (*github.IssueEvent, error) {
	events, _, err := c.c.Issues.ListIssueEvents(ctx, c.owner, c.repo, issue.GetNumber(), &github.ListOptions{PerPage: 1000})
	if err != nil {
		return nil, err
	}
	for _, e := range events {
		if e.GetEvent() == "merged" {
			return e, nil
		}
	}
	return nil, fmt.Errorf("merge event not found")
}

func (c *Client) getMergedPRs(issues []*github.Issue) (prs []*github.Issue) {
	ctx := context.Background()

	prChan := make(chan *github.Issue)

	var wg sync.WaitGroup
	for _, ii := range issues {
		if ii.PullRequestLinks == nil {
			log.Infof("%v not a pull request", issueToString(ii))
			continue
		}
		wg.Add(1)
		go func(ii *github.Issue) {
			defer wg.Done()
			// ii is a PR.
			_, err := c.getMergeEventForPR(ctx, ii)
			if err != nil {
				log.Infof("failed to get merge event, issue: %v, error: %v", issueToString(ii), err)
				return
			}
			prChan <- ii
		}(ii)
	}
	go func() {
		wg.Wait()
		close(prChan)
	}()

	for ii := range prChan {
		log.Info(issueToString(ii))
		log.Info(" - ", labelsToString(ii.Labels))
		prs = append(prs, ii)
	}
	return
}

func (c *Client) getMergedPRsForMilestone(milestoneTitle string) []*github.Issue {
	num, err := c.getMilestoneNumberForTitle(context.Background(), milestoneTitle)
	if err != nil {
		log.Info("failed to get milestone number: ", err)
	}

	// Get closed issues with milestone number.
	milestoneNumberStr := strconv.Itoa(num)
	log.Info("milestone number: ", milestoneNumberStr)
	issues, _, err := c.c.Issues.ListByRepo(context.Background(), c.owner, c.repo,
		&github.IssueListByRepoOptions{
			State:       "closed",
			Milestone:   milestoneNumberStr,
			ListOptions: github.ListOptions{PerPage: 1000},
		},
	)
	if err != nil {
		log.Info("failed to get closed issues for milestone: ", err)
		return nil
	}
	log.Info("count issues", len(issues))
	return c.getMergedPRs(issues)
}

func (c *Client) getMergedPRsForLabels(labels []string) []*github.Issue {
	// Get closed issues with labels.
	log.Info("labels: ", labels)
	issues, _, err := c.c.Issues.ListByRepo(context.Background(), c.owner, c.repo,
		&github.IssueListByRepoOptions{
			State:       "closed",
			Labels:      labels,
			ListOptions: github.ListOptions{PerPage: 1000},
		},
	)
	if err != nil {
		log.Info("failed to get closed issues for milestone: ", err)
		return nil
	}
	log.Info("count issues", len(issues))
	return c.getMergedPRs(issues)
}

func (c *Client) getOrgMembers(org string) map[string]struct{} {
	opt := &github.ListMembersOptions{}
	var count int
	ret := make(map[string]struct{})
	for {
		members, resp, err := c.c.Organizations.ListMembers(context.Background(), org, opt)
		if err != nil {
			log.Info("failed to get org members: ", err)
			return nil
		}
		for _, m := range members {
			ret[m.GetLogin()] = struct{}{}
		}
		count += len(members)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	log.Infof("%v members in org %v\n", count, org)
	return ret
}
