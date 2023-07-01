package presence

import (
	"log"
	"math/rand"
	"time"

	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/spotify"
)

type Application struct {
	spotify *spotify.Client
	github  *github.Client
	emojis  []string
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

func WithEmojis(emojis []string) Option {
	return func(app *Application) {
		app.emojis = emojis
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

		emoji := ":green_circle:"

		if len(app.emojis) > 0 {
			emoji = app.emojis[randInt(0, len(app.emojis)-1)]
		}

		err := app.github.ChangeUserStatus(
			emoji,
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

func randInt(min, max int) int {
	return rand.Intn((max+1)-min) + min
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
