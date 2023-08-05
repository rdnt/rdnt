package contributions

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	githubOauth "golang.org/x/oauth2/github"
	"gotest.tools/v3/assert"

	authn "github.com/rdnt/rdnt/pkg/oauth"
	"github.com/rdnt/rdnt/pkg/secretsmanager"
)

func TestGitPush(t *testing.T) {
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

	githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectUrl := os.Getenv("GITHUB_REDIRECT_URL")

	githubConf := &oauth2.Config{
		ClientID:     githubClientId,
		ClientSecret: githubClientSecret,
		Scopes:       []string{"user", "repo"},
		Endpoint:     githubOauth.Endpoint,
		RedirectURL:  githubRedirectUrl,
	}

	githubTokenProv, err := authn.NewMongoTokenProvider(sm, "github")
	if err != nil {
		log.Fatal(err)
	}

	githubAuthn, err := authn.NewAuthn("GraphqlClient", githubConf, githubTokenProv)
	if err != nil {
		log.Fatal(err)
	}

	c := Contributions{
		username: "rdnt",
		authn:    githubAuthn,
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

		c.String(http.StatusOK, "GraphqlClient successfully authenticated. You may close this window.")
		log.Print("GraphqlClient client authenticated successfully.")
	})

	go func() {
		err := r.Run(fmt.Sprintf("%s:%d", "localhost", 8080))
		if !errors.Is(err, http.ErrServerClosed) && err != nil {
			log.Fatal(err)
		}
	}()

	err = c.commitAndPush(context.Background())
	assert.NilError(t, err)
}
