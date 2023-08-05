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

	"github.com/rdnt/rdnt/internal/contributions"
	"github.com/rdnt/rdnt/pkg/github"
	authn "github.com/rdnt/rdnt/pkg/oauth"
	"github.com/rdnt/rdnt/pkg/secretsmanager"
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

	githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectUrl := os.Getenv("GITHUB_REDIRECT_URL")

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

	oauth := r.Group("/oauth")
	oauth.Use(gin.Logger())

	oauth.GET("/github/callback", func(c *gin.Context) {
		err := githubAuthn.ExtractToken(c.Request)
		if errors.Is(err, authn.ErrExchanged) {
			return
		} else if err != nil {
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

	githubHttpClient, err := githubAuthn.Client()
	if err != nil {
		log.Fatal(err)
	}

	githubClient := github.New(
		"https://api.github.com/graphql",
		githubHttpClient,
	)

	username := os.Getenv("USERNAME")

	p := contributions.New(contributions.Options{
		GitHub:   githubClient,
		Username: username,
	})

	err = p.Run()
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	p.Stop()
}
