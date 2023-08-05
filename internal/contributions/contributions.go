package contributions

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"

	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/graph"
	authn "github.com/rdnt/rdnt/pkg/oauth"
)

type Contributions struct {
	graphqlClient *github.Client
	httpClient    *http.Client
	username      string

	cancel context.CancelFunc
	done   chan bool
	authn  *authn.Authn
}

type Options struct {
	GraphqlClient *github.Client
	HttpClient    *http.Client
	Username      string
	TokenProvider *authn.Authn
}

func New(opts Options) *Contributions {
	c := &Contributions{
		graphqlClient: opts.GraphqlClient,
		username:      opts.Username,
		httpClient:    opts.HttpClient,
		authn:         opts.TokenProvider,
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
			fmt.Println("Querying GraphqlClient...")
			err := c.generateContributionsGraph(ctx)
			if err != nil {
				log.Println(err)
				continue
			}

			err = c.commitAndPush(ctx)
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

	contribs, err := c.graphqlClient.ContributionsView(ctx, c.username, from, to)
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

func (c *Contributions) commitAndPush(ctx context.Context) error {
	fs := osfs.New("./test")

	//f, err := fs.Create("contributions-dark.svg")
	//if err != nil {
	//	return err
	//}
	//
	//_, err = f.Write([]byte("Hello World"))
	//if err != nil {
	//	return err
	//}
	//
	//f.Close()

	r, err := git.Init(memory.NewStorage(), fs)
	if err != nil {
		return err
	}

	rem, err := r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"https://github.com/rdnt/rdnt-test"},
	})
	if err != nil {
		return err
	}

	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Add("assets")
	if err != nil {
		return err
	}

	_, err = wt.Commit("Update contributions graph", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "github-actions[bot]",
			Email: "github-actions[bot]@users.noreply.github.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("assets"),
		Create: true,
	})
	if err != nil {
		return err
	}

	tok, err := c.authn.Token()
	if err != nil {
		return err
	}

	err = rem.PushContext(ctx, &git.PushOptions{
		RefSpecs:   []config.RefSpec{"+refs/heads/assets:refs/heads/assets"},
		RemoteName: "origin",
		Force:      true,
		Auth: &githttp.BasicAuth{
			Username: "rdnt",
			Password: tok.AccessToken,
		},
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		fmt.Println("Nothing to commit.")
		return nil
	} else if err != nil {
		return err
	}

	fmt.Println("Updated.")

	return nil
}
