package google

import (
	"context"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2/google"
)

const (
	Key = "GoogleClient"
)

var clientScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
}

//go:generate counterfeiter . Client

type Client interface {
	NewToken() (string, error)
}

func NewClient() Client {
	return &client{}
}

type client struct{}

func (client) NewToken() (string, error) {
	tokenSource, err := google.DefaultTokenSource(context.Background(), clientScopes...)
	if err != nil {
		return "", err
	}

	token, err := tokenSource.Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func Instance(c *gin.Context) Client {
	return c.MustGet(Key).(Client)
}
