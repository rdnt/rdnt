package github

import (
	"context"
	"net/http"
	"time"

	_ "github.com/Khan/genqlient/generate"
	"github.com/Khan/genqlient/graphql"
	githubcontrib "github.com/rdnt/contribs-graph/github"
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
}

func (c *Client) GetContributions(ctx context.Context, user string, from, to time.Time) (githubcontrib.ContributionsResponse, error) {
	resp, err := contributionsView(ctx, c.gql, user, from, to)
	if err != nil {
		return githubcontrib.ContributionsResponse{}, err
	}

	isHaloween := resp.User.ContributionsCollection.ContributionCalendar.IsHalloween

	contribs := githubcontrib.ContributionsResponse{
		IsHalloween: isHaloween,
	}

	for _, w := range resp.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			contribs.Contributions = append(contribs.Contributions, githubcontrib.Contribution{
				Count: d.ContributionCount,
				Color: normalizeColor(d.Color, isHaloween),
			})
		}
	}

	return contribs, nil
}

func normalizeColor(color string, haloween bool) string {
	if !haloween {
		return color
	}

	switch color {
	case "#ebedf0":
		return "#ebedf0"
	case "#ffee4a":
		return "#9be9a8"
	case "#ffc501":
		return "#40c463"
	case "#fe9600":
		return "#30a14e"
	case "#03001c":
		return "#216e39"
	default:
		return "#ebedf0"
	}
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
