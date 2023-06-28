package authn

import (
	"encoding/json"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/rdnt/rdnt/pkg/secretsmanager"
)

type MongoTokenProvider struct {
	updated chan *oauth2.Token
	sm      *secretsmanager.Manager
	tokenId string
}

func NewMongoTokenProvider(sm *secretsmanager.Manager, tokenId string) (*MongoTokenProvider, error) {
	return &MongoTokenProvider{
		updated: make(chan *oauth2.Token),
		sm:      sm,
		tokenId: tokenId,
	}, nil
}

func (m *MongoTokenProvider) Get() (*oauth2.Token, error) {
	b, err := m.sm.Get(m.tokenId)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to set oauth2 token")
	}

	var token *oauth2.Token
	err = json.Unmarshal(b, &token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal oauth2 token")
	}

	return token, nil
}

func (m *MongoTokenProvider) Set(token *oauth2.Token) error {
	b, err := json.Marshal(token)
	if err != nil {
		return errors.Wrap(err, "failed to marshal oauth2 token")
	}

	err = m.sm.Set(m.tokenId, b)
	if err != nil {
		return errors.WithMessage(err, "failed to set oauth2 token")
	}

	m.updated <- token

	return nil
}

func (m *MongoTokenProvider) Updated() chan *oauth2.Token {
	return m.updated
}
