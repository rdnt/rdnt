package contributions

import (
	"context"
	"os"
	"time"

	"github.com/rdnt/contributions-graph"
	"github.com/samber/lo"

	"github.com/rdnt/rdnt/pkg/github"
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

	contribsView, err := c.graphqlClient.ContributionsView(ctx, c.username, from, to)
	if err != nil {
		return err
	}

	//stats, err := c.graphqlClient.UserStats(ctx, c.username)
	//if err != nil {
	//	return err
	//}

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

	if contribsView.IsHalloween {
		contribsView.Contributions = normalizeHalloweenColors(contribsView.Contributions)
	}

	contribs := lo.Map(contribsView.Contributions, func(c github.Contribution, _ int) graph.ContributionDay {
		return graph.ContributionDay{
			Count: c.Count,
			Color: c.Color,
		}
	})

	g := graph.NewGraph(contribs)

	err = g.Render(fd, customGithubContributionsDarkTheme)
	if err != nil {
		return err
	}

	err = g.Render(fl, customGithubContributionsLightTheme)
	if err != nil {
		return err
	}

	return nil
}

func customGithubContributionsDarkTheme(color string) string {
	switch color {
	case graph.Color0:
		color = "#2e2e3b"
	case graph.Color1:
		color = "#CAB9F8"
	case graph.Color2:
		color = "#A791E9"
	case graph.Color3:
		color = "#9473ee"
	case graph.Color4:
		color = "#7347ea"
	}
	return color
}

func customGithubContributionsLightTheme(color string) string {
	switch color {
	case graph.Color0:
		color = "#F6F3FF"
	case graph.Color1:
		color = "#E4DBFC"
	case graph.Color2:
		color = "#CAB9F8"
	case graph.Color3:
		color = "#B8A2F6"
	case graph.Color4:
		color = "#AA8FF5"
	}
	return color
}

func normalizeHalloweenColors(contribs []github.Contribution) []github.Contribution {
	for i, c := range contribs {
		var color string
		switch c.Color {
		case "#ebedf0":
			color = "#ebedf0"
		case "#ffee4a":
			color = "#9be9a8"
		case "#ffc501":
			color = "#40c463"
		case "#fe9600":
			color = "#30a14e"
		case "#03001c":
			color = "#216e39"
		default:
			color = "#ebedf0"
		}
		contribs[i].Color = color
	}

	return contribs
}
