package main

import (
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teambition/rrule-go"
)

func TestGHClient(t *testing.T) {
	gh := ghClient("TestOrg/TestRepo", "abc123")
	assert.Equal(t, gh.owner, "TestOrg")
	assert.Equal(t, gh.repo, "TestRepo")
	assert.NotNil(t, gh.ctx)
	// *github.Client does not expose its underlying *http.Client, so it's not
	// possible to assert that it's using an *oauth2.Transport with the
	// appropriate token.
}

func TestGetTimes(t *testing.T) {
	rr, err := rrule.StrToRRule("FREQ=WEEKLY;BYDAY=FR;DTSTART=20191211T220000Z")
	count := 4

	times, err := getTimes(rr, uint8(count))
	require.NoError(t, err)

	assert.Len(t, times, count)
}

func TestDoTheThing(t *testing.T) {
	var (
		client   = new(testClient)
		format   = "Due 2006-Jan-02"
		times    []time.Time
		validate = func(err error) {
			require.NoError(t, err)
			idx := sliceToIdx(client.milestones)
			assert.Len(t, idx, len(times)) // assert no dupes, no nil Titles
			for _, ti := range times {
				_, ok := idx[ti.Format(format)]
				assert.True(t, ok)
			}
		}
	)

	// no err for no-op
	validate(doTheThing(client, format, times))

	// creates one time just fine
	times = append(times, time.Date(2019, time.December, 13, 22, 00, 00, 0000, time.UTC))
	validate(doTheThing(client, format, times))

	// creates second time just fine
	times = append(times, time.Date(2019, time.December, 20, 22, 00, 00, 0000, time.UTC))
	validate(doTheThing(client, format, times))

	// does not re-create when non-title attribute is modified
	{
		newDueDate := time.Date(2019, time.December, 19, 22, 00, 00, 0000, time.UTC)
		*client.milestones[1].DueOn = newDueDate
	}
	validate(doTheThing(client, format, times))

	// is idempotent
	validate(doTheThing(client, format, times))
}

var _ ifi = &testClient{}

type testClient struct {
	milestones []*github.Milestone
}

func (c *testClient) ListMilestones(_ *github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error) {
	return c.milestones, new(github.Response), nil
}

func (c *testClient) CreateMilestone(m *github.Milestone) (*github.Milestone, *github.Response, error) {
	c.milestones = append(c.milestones, m)
	return m, new(github.Response), nil
}

func TestSliceToIdx(t *testing.T) {
	strPtr := func(str string) *string { return &str }
	m1 := &github.Milestone{Title: strPtr("Foo")}
	m2 := &github.Milestone{Title: strPtr("Bar")}
	m3 := &github.Milestone{Title: strPtr("Baz")}
	m4 := &github.Milestone{}
	m5 := &github.Milestone{Title: strPtr("Foo")}
	s := []*github.Milestone{m1, m2, m3, m4, m5}
	m := map[string]*github.Milestone{
		"Foo": m1,
		"Bar": m2,
		"Baz": m3,
	}

	assert.Equal(t, sliceToIdx(s), m)
}
