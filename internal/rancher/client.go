package rancher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type NewTokenRequest struct {
	ResponseType string `json:"responseType"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

var (
	errNotFoundFormat = "error getting token: %s"
)

func NewClient() *Client {
	return &Client{
		c: &http.Client{},
	}
}

type Client struct {
	c           *http.Client
	cachedToken KubeconfigToken
	mux         sync.Mutex
	password    string
	timeout     time.Duration
	url         string
	username    string
}

func (c *Client) Token(ctx context.Context) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if time.Now().In(time.UTC).After(c.cachedToken.ExpiresAt) || c.cachedToken.Token == "" {
		k := KubeconfigToken{}

		data := NewTokenRequest{
			ResponseType: "kubeconfig",
			Username:     c.username,
			Password:     c.password,
		}

		b, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		// Configure request to time out.
		ctx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()
		// Create the request.
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewBuffer(b))
		if err != nil {
			return "", err
		}

		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")

		res, err := c.c.Do(req)
		if err != nil {
			return "", err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			_, _ = io.Copy(ioutil.Discard, res.Body)

			return "", fmt.Errorf(errNotFoundFormat, res.Status)
		}

		err = json.NewDecoder(res.Body).Decode(&k)
		if err != nil {
			return "", err
		}

		c.cachedToken = k
	}

	return c.cachedToken.Token, nil
}

// WithPassword sets the password.
func (c *Client) WithPassword(password string) {
	c.password = password
}

// WithTransport sets the http transport.
func (c *Client) WithTransport(transport *http.Transport) {
	c.c.Transport = transport
}

// WithTimeout sets the timeout on the http request to retrieve the token.
func (c *Client) WithTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// WithURL sets the URL.
func (c *Client) WithURL(url string) {
	c.url = url
}

// WithUsername sets the username.
func (c *Client) WithUsername(username string) {
	c.username = username
}
