package google

import (
	"context"

	"golang.org/x/oauth2/google"
)

var clientScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
}

func NewToken() (string, error) {
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
