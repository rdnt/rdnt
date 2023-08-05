package contributions

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/graph"
)

type Contributions struct {
	github   *github.Client
	username string

	cancel context.CancelFunc
	done   chan bool
}

type Options struct {
	GitHub   *github.Client
	Username string
}

func New(opts Options) *Contributions {
	c := &Contributions{
		github:   opts.GitHub,
		username: opts.Username,
	}

	return c
}

func (c *Contributions) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.done = make(chan bool)

	go c.githubUpdateLoop(ctx)

	log.Print("Contributions started.")

	return nil
}

func (c *Contributions) Stop() {
	c.cancel()
	<-c.done
	close(c.done)
	c.done = nil
	c.cancel = nil

	log.Print("Contributions stopped.")
}

func (c *Contributions) githubUpdateLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	defer func() {
		ticker.Stop()
		c.done <- true
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("Querying GitHub...")
			err := c.generateContributionsGraph(ctx)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func (c *Contributions) generateContributionsGraph(ctx context.Context) error {
	from := time.Now().UTC().AddDate(-1, 0, -7)
	to := time.Now().UTC()

	contribs, err := c.github.ContributionsView(ctx, c.username, from, to)
	if err != nil {
		return err
	}

	err = os.Mkdir("assets", os.ModePerm)
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

	g := graph.NewGraph(contribs)

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
