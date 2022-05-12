package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/graph"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	username := os.Getenv("USERNAME")
	accessToken := os.Getenv("ACCESS_TOKEN")

	contribs, err := github.ContributionsPerDay(ctx, username, accessToken)
	if err != nil {
		log.Println(err)
		return
	}
	
	err = os.Mkdir("assets", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Println(err)
		return
	}

	fd, err := os.Create("assets/contributions-dark.svg")
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		_ = fd.Close()
	}()

	fl, err := os.Create("assets/contributions-light.svg")
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		_ = fl.Close()
	}()

	g := graph.NewGraph(contribs)

	err = g.Render(fd, graph.Dark)
	if err != nil {
		log.Println(err)
		return
	}

	err = g.Render(fl, graph.Light)
	if err != nil {
		log.Println(err)
		return
	}
}
