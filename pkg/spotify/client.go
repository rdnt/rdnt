package spotify

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
	"net/http"
	"reflect"
	"strings"
)

type Track struct {
	Track     string
	Artist    string
	Thumbnail string
}

type TrackChangedHandler func(state *Track)

type Client struct {
	client *spotify.Client

	track        *Track
	TrackChanged TrackChangedHandler
}

func New(httpClient *http.Client) *Client {
	return &Client{
		client: spotify.New(
			httpClient,
			spotify.WithRetry(true),
		),
	}
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

	if c.TrackChanged != nil {
		c.TrackChanged(c.track)
	}

	return nil
}
func (c *Client) getCurrentTrack() (*Track, error) {
	if c.client == nil {
		return nil, errors.New("not authenticated")
	}

	res, err := c.client.PlayerCurrentlyPlaying(context.Background())
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
