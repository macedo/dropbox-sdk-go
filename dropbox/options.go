package dropbox

import "net/http"

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Option func(*Client)

func WithHTTPClient(cli HTTPClient) Option {
	return func(c *Client) {
		c.HTTPClient = cli
	}
}
