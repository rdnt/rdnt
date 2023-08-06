package presence

import (
	"context"
	"log"
	"time"

	"github.com/rdnt/rdnt/pkg/github"
	"github.com/rdnt/rdnt/pkg/rand"
	"github.com/rdnt/rdnt/pkg/spotify"
)

type Presence struct {
	spotify *spotify.Client
	github  *github.Client
	emojis  []string

	cancel context.CancelFunc
	done   chan bool
}

type Options struct {
	Spotify *spotify.Client
	GitHub  *github.Client
	Emojis  []string
}

func New(opts Options) *Presence {
	p := &Presence{
		spotify: opts.Spotify,
		github:  opts.GitHub,
		emojis:  opts.Emojis,
	}

	return p
}

func (p *Presence) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.done = make(chan bool)

	go p.githubUpdateLoop(ctx)
	go p.spotifyPollLoop(ctx)

	log.Print("Presence started.")

	return nil
}

func (p *Presence) Stop() {
	p.cancel()
	<-p.done
	<-p.done
	close(p.done)
	p.done = nil
	p.cancel = nil

	log.Print("Presence stopped.")
}

func (p *Presence) githubUpdateLoop(ctx context.Context) {
	trackChan, dispose := p.spotify.Track()

	defer func() {
		dispose()
		p.done <- true
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case track := <-trackChan:
			if track == nil {
				err := p.github.ChangeUserStatus(ctx, "", time.Time{}, "", false)
				if err != nil {
					log.Print(err)
					continue
				}

				log.Printf("Status cleared.")
				continue
			}

			status := "Listening to " + track.Track + " - " + track.Artist

			emoji := ":green_circle:"

			if len(p.emojis) > 0 && p.emojis[0] != "" {
				emoji = p.emojis[rand.Int(0, len(p.emojis)-1)]
			}

			err := p.github.ChangeUserStatus(ctx, emoji,
				time.Now().UTC().Add(120*time.Minute),
				status, false,
			)
			if err != nil {
				log.Print(err)
				continue
			}

			log.Printf("Status updated to: \"%s\".", status)
		}
	}
}

func (p *Presence) spotifyPollLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	defer func() {
		ticker.Stop()
		p.done <- true
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := p.spotify.UpdateCurrentTrack()
			if err != nil {
				log.Print(err)
				continue
			}
		}
	}
}
