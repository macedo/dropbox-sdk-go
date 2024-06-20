package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	AuthenticationEndpoint = "https://api.dropbox.com"
)

const ServiceAPIVersion = "2"

type Client struct {
	AuthenticationEndpoint string
	ContentUploadEndpoint  string
	HTTPClient             HTTPClient
}

func New(opts ...Option) *Client {
	cli := &Client{
		AuthenticationEndpoint: AuthenticationEndpoint,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	cli.parseOptions(opts...)
	return cli
}

func (c *Client) parseOptions(opts ...Option) {
	for _, option := range opts {
		option(c)
	}
}

func (c *Client) sendRequest(r *http.Request, v any) error {
	response, err := c.HTTPClient.Do(r)
	if err != nil {
		return fmt.Errorf("HTTPClient.Do: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		b, err := io.ReadAll(response.Body)
		if err != nil {
			return nil
		}

		return fmt.Errorf("request failed: %q", string(b))
	}

	return json.NewDecoder(response.Body).Decode(&v)
}
