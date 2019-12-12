package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/teambition/rrule-go"
	"golang.org/x/oauth2"
)

type config struct {
	Recurrence *rule  `required:"true"`
	Format     string `default:"2006-01-02"`
	Count      uint8  `default:"4"`
	Token      string `envconfig:"GITHUB_TOKEN" required:"true"`
	Repository string `envconfig:"GITHUB_REPOSITORY" required:"true"`
}

type rule struct {
	*rrule.RRule
}

func (r *rule) Set(value string) error {
	rr, err := rrule.StrToRRule(value)
	r.RRule = rr
	return err
}

func main() {
	var c config
	if err := envconfig.Process("INPUT", &c); err != nil {
		log.Fatal(err)
	}

	times, err := getTimes(c.Recurrence.RRule, c.Count)
	if err != nil {
		log.Fatal(err)
	}

	client := ghClient(c.Repository, c.Token)
	if err := doTheThing(client, c.Format, times); err != nil {
		log.Fatal(err)
	}
}

func getTimes(rr *rrule.RRule, count uint8) ([]time.Time, error) {
	next := rr.After(time.Now(), true)
	var zero time.Time
	if next == zero {
		return nil, errors.New("No more occurrences exist")
	}
	rr.DTStart(next)

	iter := rr.Iterator()
	times := make([]time.Time, count)
	for i := range times {
		value, ok := iter()
		if !ok {
			return nil, errors.New("No more occurrences exist")
		}
		times[i] = value
	}
	return times, nil
}

func ghClient(repo, token string) *gh {
	parts := strings.SplitN(repo, "/", 2)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	return &gh{
		Client: github.NewClient(tc),
		owner:  parts[0],
		repo:   parts[1],
		ctx:    ctx,
	}
}

func doTheThing(client ifi, format string, times []time.Time) error {
	milestones, _, err := client.ListMilestones(&github.MilestoneListOptions{
		State:     "all",
		Direction: "desc",
	})
	if err != nil {
		return errors.Wrap(err, "could not list milestones")
	}

	idx := sliceToIdx(milestones)
	for _, t := range times {
		name := t.Format(format)
		if _, ok := idx[name]; ok {
			continue
		}

		if _, _, err := client.CreateMilestone(&github.Milestone{
			Title: &name,
			DueOn: &t,
		}); err != nil {
			return errors.Wrap(err, "could not create milestone")
		}
	}

	return nil
}

type ifi interface {
	ListMilestones(*github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error)
	CreateMilestone(*github.Milestone) (*github.Milestone, *github.Response, error)
}

type gh struct {
	*github.Client
	owner, repo string
	ctx         context.Context
}

func (c *gh) ListMilestones(opts *github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error) {
	return c.Issues.ListMilestones(c.ctx, c.owner, c.repo, opts)
}

func (c *gh) CreateMilestone(m *github.Milestone) (*github.Milestone, *github.Response, error) {
	return c.Issues.CreateMilestone(c.ctx, c.owner, c.repo, m)
}

func sliceToIdx(s []*github.Milestone) map[string]*github.Milestone {
	idx := map[string]*github.Milestone{}
	for _, m := range s {
		if m.Title == nil {
			continue
		}
		idx[*m.Title] = m
	}
	return idx
}
