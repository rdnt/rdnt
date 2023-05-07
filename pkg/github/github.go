package github

import (
	"context"
	"github.com/Khan/genqlient/graphql"
	"net/http"
	"time"
)

// imports needed for genqlient
import (
	_ "github.com/Khan/genqlient/generate"
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

func (c *Client) ChangeUserStatus(emoji string, expiresAt time.Time, message string, limited bool) error {
	_, err := changeUserStatus(context.Background(), c.gql, ChangeUserStatusInput{
		ClientMutationId:    &c.clientId,
		Emoji:               &emoji,
		ExpiresAt:           &expiresAt,
		LimitedAvailability: &limited,
		Message:             &message,
		OrganizationId:      nil,
	})

	return err
}
