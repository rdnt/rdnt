package status

import (
	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/spotify"
	"log"
	"time"
)

type Application struct {
	spotify *spotify.Client
	github  *github.Client
}

type Option func(app *Application)

func WithSpotifyClient(c *spotify.Client) Option {
	return func(app *Application) {
		app.spotify = c
	}
}

func WithGithubClient(c *github.Client) Option {
	return func(app *Application) {
		app.github = c
	}
}

func New(opts ...Option) *Application {
	app := &Application{}

	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}

	app.spotify.TrackChanged = func(track *spotify.Track) {
		if track == nil {
			err := app.github.ChangeUserStatus("", time.Time{}, "", false)
			if err != nil {
				log.Print(err)
				return
			}

			log.Printf("Status cleared.")
			return
		}

		status := "Listening to " + track.Track + " - " + track.Artist

		err := app.github.ChangeUserStatus(
			":green_circle:",
			time.Now().Add(120*time.Minute).UTC(), // listening to a 2-hour monstercat mix? plausible
			status,
			true,
		)
		if err != nil {
			log.Print(err)
			return
		}

		log.Printf("Status updated to: \"%s\".", status)
	}

	return app
}

func (app *Application) Start() error {
	go func() {
		for {
			func() {
				defer time.Sleep(1 * time.Second)

				err := app.spotify.UpdateCurrentTrack()
				if err != nil {
					log.Print(err)
				}
			}()
		}
	}()

	log.Print("App started.")

	return nil
}
