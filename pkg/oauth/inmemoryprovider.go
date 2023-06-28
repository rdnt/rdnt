package authn

import "golang.org/x/oauth2"

type InMemoryTokenProvider struct {
	token   *oauth2.Token
	updated chan *oauth2.Token
}

func NewInMemoryTokenProvider() *InMemoryTokenProvider {
	return &InMemoryTokenProvider{updated: make(chan *oauth2.Token)}
}

func (t *InMemoryTokenProvider) Get() (*oauth2.Token, error) {
	return t.token, nil
}

func (t *InMemoryTokenProvider) Set(token *oauth2.Token) error {
	t.token = token
	t.updated <- t.token

	return nil
}

func (t *InMemoryTokenProvider) Updated() chan *oauth2.Token {
	return t.updated
}
