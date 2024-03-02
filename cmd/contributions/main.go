package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rdnt/contribs-graph/graph"
	"golang.org/x/oauth2"

	githubcontrib "github.com/rdnt/contribs-graph/github"

	"github.com/rdnt/rdnt/pkg/github"
)

func main() {
	accessToken := os.Getenv("ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("ACCESS_TOKEN is required")
	}

	username := os.Getenv("USERNAME")
	if username == "" {
		log.Fatal("USERNAME is required")
	}

	tokSrc := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})

	ghHttpClient := oauth2.NewClient(context.Background(), tokSrc)

	githubClient := github.New(
		"https://api.github.com/graphql",
		ghHttpClient,
	)

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	contribs, err := githubcontrib.Contributions(reqCtx, githubClient, username)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("assets", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	g := graph.New(contribs)

	fd, err := os.Create("assets/contributions-dark.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	err = g.Render(fd, customDarkTheme)
	if err != nil {
		log.Fatal(err)
	}

	fl, err := os.Create("assets/contributions-light.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer fl.Close()

	err = g.Render(fl, customLightTheme)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Contribution graphs generated.")
}

func customDarkTheme(color string) string {
	switch color {
	case graph.Color0:
		color = "#22262d"
	case graph.Color1:
		color = "#7347ea"
	case graph.Color2:
		color = "#9473ee"
	case graph.Color3:
		color = "#A791E9"
	case graph.Color4:
		color = "#CAB9F8"
	}
	return color
}

func customLightTheme(color string) string {
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
