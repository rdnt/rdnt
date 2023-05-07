package authn

import "golang.org/x/oauth2"

type TokenProvider struct {
	token   *oauth2.Token
	updated chan *oauth2.Token
}

func NewTokenProvider() *TokenProvider {
	return &TokenProvider{updated: make(chan *oauth2.Token)}
}

func (t *TokenProvider) Get() (*oauth2.Token, error) {
	return t.token, nil
}

func (t *TokenProvider) Set(token *oauth2.Token) error {
	t.token = token
	t.updated <- t.token

	return nil
}

func (t *TokenProvider) Updated() chan *oauth2.Token {
	return t.updated
}
