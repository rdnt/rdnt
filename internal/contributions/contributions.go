package contributions

import (
	"context"
	"os"
	"time"

	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/graph"
	authn "github.com/rdnt/rdnt/pkg/oauth"
)

type Contributions struct {
	graphqlClient *github.Client
	username      string

	cancel context.CancelFunc
	done   chan bool
	authn  *authn.Authn
}

type Options struct {
	GraphqlClient *github.Client
	Username      string
}

func New(opts Options) *Contributions {
	c := &Contributions{
		graphqlClient: opts.GraphqlClient,
		username:      opts.Username,
	}

	return c
}

func (c *Contributions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := c.generateContributionsGraph(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *Contributions) generateContributionsGraph(ctx context.Context) error {
	from := time.Now().UTC().AddDate(-1, 0, -7)
	to := time.Now().UTC()

	contribs, err := c.graphqlClient.ContributionsView(ctx, c.username, from, to)
	if err != nil {
		return err
	}

	stats, err := c.graphqlClient.UserStats(ctx, c.username)
	if err != nil {
		return err
	}

	err = os.MkdirAll("assets", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	fd, err := os.Create("assets/contributions-dark.svg")
	if err != nil {
		return err
	}
	defer func() {
		_ = fd.Close()
	}()

	fl, err := os.Create("assets/contributions-light.svg")
	if err != nil {
		return err
	}
	defer func() {
		_ = fl.Close()
	}()

	g := graph.NewGraph(contribs, stats)

	err = g.Render(fd, graph.Dark)
	if err != nil {
		return err
	}

	err = g.Render(fl, graph.Light)
	if err != nil {
		return err
	}

	return nil
}
