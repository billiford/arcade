package google

import (
	"context"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
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
	NewToken() (*oauth2.Token, error)
}

func NewClient() Client {
	return &client{}
}

type client struct{}

func (client) NewToken() (*oauth2.Token, error) {
	tokenSource, err := google.DefaultTokenSource(context.Background(), clientScopes...)
	if err != nil {
		return nil, err
	}

	token, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func Instance(c *gin.Context) Client {
	return c.MustGet(Key).(Client)
}
