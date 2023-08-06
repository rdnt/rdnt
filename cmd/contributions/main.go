package main

import (
	"context"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/rdnt/rdnt/internal/contributions"
	"github.com/rdnt/rdnt/pkg/github"
)

func main() {
	accessToken := os.Getenv("ACCESS_TOKEN")
	tokSrc := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	githubHttpClient := oauth2.NewClient(ctx, tokSrc)

	githubClient := github.New(
		"https://api.github.com/graphql",
		githubHttpClient,
	)

	username := os.Getenv("USERNAME")

	c := contributions.New(contributions.Options{
		GraphqlClient: githubClient,
		Username:      username,
	})

	err := c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
