package spotify

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"

	"github.com/rdnt/rdnt/pkg/broker"
)

type Track struct {
	Track     string
	Artist    string
	Thumbnail string
}

type TrackChangedHandler func(state *Track)

type Client struct {
	client      *spotify.Client
	track       *Track
	trackBroker broker.Broker[*Track]
}

func New(httpClient *http.Client) *Client {
	return &Client{
		client: spotify.New(
			httpClient,
			spotify.WithRetry(true),
		),
		trackBroker: broker.New[*Track](),
	}
}

func (c *Client) Track() (track <-chan *Track, dispose func()) {
	return c.trackBroker.Subscribe()
}

func (c *Client) UpdateCurrentTrack() error {
	track, err := c.getCurrentTrack()
	if err != nil {
		return err
	}

	if reflect.DeepEqual(c.track, track) {
		// no change
		return nil
	}

	c.track = track
	c.trackBroker.Publish(c.track)

	return nil
}
func (c *Client) getCurrentTrack() (*Track, error) {
	if c.client == nil {
		return nil, errors.New("not authenticated")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.client.PlayerCurrentlyPlaying(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current track")
	}

	if res.Item == nil || !res.Playing {
		// nothing playing
		return nil, nil
	}

	artists := make([]string, len(res.Item.Artists))
	for i, artist := range res.Item.Artists {
		artists[i] = artist.Name
	}

	return &Track{
		Track:     res.Item.Name,
		Artist:    strings.Join(artists, ", "),
		Thumbnail: res.Item.Album.Images[0].URL,
	}, nil
}
