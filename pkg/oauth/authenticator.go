package authn

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/rdnt/rdnt/pkg/rand"
)

type Authn struct {
	name         string
	conf         *oauth2.Config
	state        string
	provider     TokenProvider
	client       *http.Client
	exchanged    bool
	exchangedMux sync.Mutex
}

var ErrTokenNotSet = errors.New("token not set")
var ErrExchanged = errors.New("token already exchanged")

type TokenProvider interface {
	Get() (*oauth2.Token, error)
	Set(*oauth2.Token) error
	Updated() chan *oauth2.Token
}

func NewAuthn(name string, cfg *oauth2.Config, prov TokenProvider) (*Authn, error) {
	state, err := rand.String(32)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate state")
	}

	return &Authn{
		name:     name,
		conf:     cfg,
		state:    state,
		provider: prov,
	}, nil
}

func (a *Authn) Client() (*http.Client, error) {
	tok, err := a.provider.Get()
	if errors.Is(err, ErrTokenNotSet) {
		go a.startOauthFlow()

		tok = <-a.provider.Updated()
		if tok == nil {
			return nil, errors.New("token not set")
		}
	} else if err != nil {
		return nil, err
	}

	return a.createClient(tok)
}

func (a *Authn) Token() (*oauth2.Token, error) {
	tok, err := a.provider.Get()
	if errors.Is(err, ErrTokenNotSet) {
		go a.startOauthFlow()

		tok = <-a.provider.Updated()
		if tok == nil {
			return nil, errors.New("token not set")
		}
	} else if err != nil {
		return nil, err
	}

	return tok, nil
}

func (a *Authn) startOauthFlow() {
	log.Printf("%s requires authentication: %s", a.name, a.oauthUrl())
}

func (a *Authn) createClient(token *oauth2.Token) (*http.Client, error) {
	if token == nil {
		return nil, errors.New("invalid token")
	}

	return a.conf.Client(context.Background(), token), nil
}

func (a *Authn) oauthUrl() string {
	return a.conf.AuthCodeURL(a.state, oauth2.AccessTypeOffline)
}

func (a *Authn) ExtractToken(req *http.Request) error {
	a.exchangedMux.Lock()
	if a.exchanged {
		a.exchangedMux.Unlock()
		return ErrExchanged
	}

	a.exchanged = true
	a.exchangedMux.Unlock()

	values := req.URL.Query()
	if e := values.Get("error"); e != "" {
		return errors.New("auth failed")
	}

	code := values.Get("code")
	if code == "" {
		return errors.New("invalid access code")
	}

	actualState := values.Get("state")
	if actualState != a.state {
		return errors.New("invalid state")
	}

	tok, err := a.conf.Exchange(context.Background(), code)
	if err != nil {
		return errors.Wrap(err, "token exchange failed")
	}

	return a.provider.Set(tok)
}
