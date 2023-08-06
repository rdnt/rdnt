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

type ContributionsEntry struct {
	Count int
	Color string
}

func (c *Client) ContributionsView(ctx context.Context,
	username string, from, to time.Time,
) ([]ContributionsEntry, error) {
	resp, err := contributionsView(ctx, c.gql, username, from, to)
	if err != nil {
		return nil, err
	}

	var contributions []ContributionsEntry

	for _, w := range resp.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			contributions = append(contributions, ContributionsEntry{
				Count: d.ContributionCount,
				Color: d.Color,
			})
		}
	}

	return contributions, nil
}
