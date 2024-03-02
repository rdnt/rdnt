package github

import (
	"context"
	"net/http"
	"time"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
	_ "github.com/agnivade/levenshtein"
)

//go:generate go run github.com/Khan/genqlient genql.yaml

type Client struct {
	apiURL   string
	gql      graphql.Client
	clientId string
}

func New(apiURL string, httpClient *http.Client) *Client {
	return &Client{
		apiURL: apiURL,
		gql:    graphql.NewClient(apiURL, httpClient),
	}
}

func (c *Client) ChangeUserStatus(ctx context.Context,
	emoji string, expiresAt time.Time, message string, limited bool) error {
	_, err := changeUserStatus(ctx, c.gql, ChangeUserStatusInput{
		ClientMutationId:    &c.clientId,
		Emoji:               &emoji,
		ExpiresAt:           &expiresAt,
		LimitedAvailability: &limited,
		Message:             &message,
		OrganizationId:      nil,
	})

	return err
}

type Contributions struct {
	IsHalloween   bool
	Contributions []Contribution
	Commits       int
	Issues        int
	PullRequests  int
	Reviews       int
	Total         int
}

type Contribution struct {
	Count int
	Color string
	Date  string
}

func (c *Client) ContributionsView(ctx context.Context,
	username string, from, to time.Time,
) (Contributions, error) {
	resp, err := contributionsView(ctx, c.gql, username, from, to)
	if err != nil {
		return Contributions{}, err
	}

	contributions := Contributions{
		IsHalloween: resp.User.ContributionsCollection.ContributionCalendar.IsHalloween,
		//Commits:      resp.User.ContributionsCollection.TotalCommitContributions,
		//Issues:       resp.User.ContributionsCollection.TotalIssueContributions,
		//PullRequests: resp.User.ContributionsCollection.TotalPullRequestContributions,
		//Reviews:      resp.User.ContributionsCollection.TotalPullRequestReviewContributions,
		//Total:        resp.User.ContributionsCollection.ContributionCalendar.TotalContributions,
	}

	for _, w := range resp.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			contributions.Contributions = append(contributions.Contributions, Contribution{
				Count: d.ContributionCount,
				Color: d.Color,
			})
		}
	}

	return contributions, nil
}

type Stats struct {
	TotalContribs   int
	PrivateContribs int
	Issues          int
	PullRequests    int
}

func (c *Client) UserStats(ctx context.Context,
	username string,
) (Stats, error) {
	resp, err := userInfo(ctx, c.gql, username)
	if err != nil {
		return Stats{}, err
	}

	stats := Stats{
		//TotalContribs:   resp.User.ContributionsCollection.TotalCommitContributions,
		//PrivateContribs: resp.User.ContributionsCollection.RestrictedContributionsCount,
		PullRequests: resp.User.PullRequests.TotalCount,
		Issues:       resp.User.Issues.TotalCount,
	}

	return stats, nil
}
