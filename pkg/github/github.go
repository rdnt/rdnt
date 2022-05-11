package github

import (
	"context"
	"os"
	"time"

	githubql "github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type Contribution struct {
	Count int
}

func Contributions(ctx context.Context) ([]Contribution, error) {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("ACCESS_TOKEN")},
	)
	httpClient := oauth2.NewClient(ctx, src)
	c := githubql.NewClient(httpClient)

	var query struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					Weeks []struct {
						ContributionDays []struct {
							ContributionCount int
						}
					}
				}
			} `graphql:"contributionsCollection(from: $from, to: $to)"`
		} `graphql:"user(login: $login)"`
	}

	from := time.Now().AddDate(-1, 0, -7).In(time.UTC)
	to := time.Now().In(time.UTC)

	v := map[string]interface{}{
		"login": githubql.String(os.Getenv("USERNAME")),
		"from":  githubql.DateTime{Time: from},
		"to":    githubql.DateTime{Time: to},
	}

	err := c.Query(ctx, &query, v)
	if err != nil {
		return nil, err
	}

	var contributions []Contribution

	for _, w := range query.User.ContributionsCollection.ContributionCalendar.Weeks {
		for _, d := range w.ContributionDays {
			contributions = append(contributions, Contribution{
				Count: d.ContributionCount,
			})
		}
	}

	return contributions, nil
}
