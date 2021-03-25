package rancher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	Key = "RancherClient"
)

//go:generate counterfeiter . Client

type Client interface {
	NewToken(context.Context) (KubeconfigToken, error)
	WithURL(string)
	WithUsername(string)
	WithPassword(string)
	WithTransport(*http.Transport)
}

type NewTokenRequest struct {
	ResponseType string `json:"responseType"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

var (
	errNotFoundFormat = "error getting token: %s"
)

func NewClient() Client {
	return &client{
		c: &http.Client{},
	}
}

type client struct {
	url      string
	username string
	password string
	c        *http.Client
}

func (c *client) WithURL(url string) {
	c.url = url
}

func (c *client) WithUsername(username string) {
	c.username = username
}

func (c *client) WithPassword(password string) {
	c.password = password
}

func (c *client) WithTransport(transport *http.Transport) {
	c.c.Transport = transport
}

func (c *client) NewToken(ctx context.Context) (KubeconfigToken, error) {
	k := KubeconfigToken{}

	data := NewTokenRequest{
		ResponseType: "kubeconfig",
		Username:     c.username,
		Password:     c.password,
	}

	b, err := json.Marshal(data)
	if err != nil {
		return k, err
	}

	req, err := http.NewRequest(http.MethodPost, c.url, bytes.NewBuffer(b))
	if err != nil {
		return k, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	res, err := c.c.Do(req)
	if err != nil {
		return k, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		_, _ = io.Copy(ioutil.Discard, res.Body)
		return k, fmt.Errorf(errNotFoundFormat, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&k)
	if err != nil {
		return k, err
	}

	return k, nil
}

func Instance(c *gin.Context) Client {
	instance, exists := c.Get(Key)
	if exists {
		return instance.(Client)
	}
	return nil
}
