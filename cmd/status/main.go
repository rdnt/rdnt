package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	githubOauth "golang.org/x/oauth2/github"
	spotifyOauth "golang.org/x/oauth2/spotify"

	"github.com/rdnt/rdnt/internal/status"
	"github.com/rdnt/rdnt/pkg/github"
	authn "github.com/rdnt/rdnt/pkg/oauth"
	"github.com/rdnt/rdnt/pkg/secretsmanager"
	"github.com/rdnt/rdnt/pkg/spotify"
)

func main() {
	host := os.Getenv("SERVER_HOST")
	strPort := os.Getenv("SERVER_PORT")

	if host == "" {
		host = "localhost"
	}

	if strPort == "" {
		strPort = "8080"
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		log.Fatal(err)
	}

	spotifyClientId := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	spotifyRedirectUrl := os.Getenv("SPOTIFY_REDIRECT_URL")

	githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectUrl := os.Getenv("GITHUB_REDIRECT_URL")

	spotifyConf := &oauth2.Config{
		ClientID:     spotifyClientId,
		ClientSecret: spotifyClientSecret,
		Scopes:       []string{"user-read-private", "user-read-playback-state"},
		Endpoint:     spotifyOauth.Endpoint,
		RedirectURL:  spotifyRedirectUrl,
	}

	githubConf := &oauth2.Config{
		ClientID:     githubClientId,
		ClientSecret: githubClientSecret,
		Scopes:       []string{"user"},
		Endpoint:     githubOauth.Endpoint,
		RedirectURL:  githubRedirectUrl,
	}

	mongoAddress := os.Getenv("MONGO_ADDRESS")
	mongoDatabase := os.Getenv("MONGO_DATABASE")
	encKeyBase64 := os.Getenv("SECRETS_ENCRYPTION_KEY")
	signingKeyBase64 := os.Getenv("SECRETS_SIGNING_KEY")

	encryptionKey, err := base64.StdEncoding.DecodeString(encKeyBase64)
	if err != nil {
		log.Fatal(err)
	}

	signingKey, err := base64.StdEncoding.DecodeString(signingKeyBase64)
	if err != nil {
		log.Fatal(err)
	}

	sm, err := secretsmanager.New(mongoAddress, mongoDatabase, encryptionKey, signingKey)
	if err != nil {
		log.Fatal(err)
	}

	spotifyTokenProv, err := authn.NewMongoTokenProvider(sm, "spotify")
	if err != nil {
		log.Fatal(err)
	}

	spotifyAuthn, err := authn.NewAuthn("Spotify", spotifyConf, spotifyTokenProv)
	if err != nil {
		log.Fatal(err)
	}

	githubTokenProv, err := authn.NewMongoTokenProvider(sm, "github")
	if err != nil {
		log.Fatal(err)
	}

	githubAuthn, err := authn.NewAuthn("GitHub", githubConf, githubTokenProv)
	if err != nil {
		log.Fatal(err)
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	err = r.SetTrustedProxies(nil)
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	r.GET("/oauth/spotify/callback", func(c *gin.Context) {
		err := spotifyAuthn.ExtractToken(c.Request)
		if err != nil {
			log.Println("spotify", err)
			return
		}

		c.String(http.StatusOK, "Spotify successfully authenticated. You may close this window.")
		log.Print("Spotify client authenticated successfully.")
	})

	r.GET("/oauth/github/callback", func(c *gin.Context) {
		err := githubAuthn.ExtractToken(c.Request)
		if err != nil {
			log.Println("github", err)
			return
		}

		c.String(http.StatusOK, "GitHub successfully authenticated. You may close this window.")
		log.Print("GitHub client authenticated successfully.")
	})

	go func() {
		err := r.Run(fmt.Sprintf("%s:%d", host, port))
		if !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Fatal(err)
		}
	}()

	spotifyHttpClient, err := spotifyAuthn.Client()
	if err != nil {
		log.Fatal(err)
	}

	githubHttpClient, err := githubAuthn.Client()
	if err != nil {
		log.Fatal(err)
	}

	spotifyClient := spotify.New(spotifyHttpClient)
	if err != nil {
		log.Fatal(err)
	}

	githubClient := github.New(
		"https://api.github.com/graphql",
		githubHttpClient,
	)

	app := status.New(
		status.WithSpotifyClient(spotifyClient),
		status.WithGithubClient(githubClient),
	)

	err = app.Start()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
}
