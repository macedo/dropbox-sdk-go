package dropbox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	AuthenticationEndpoint = "https://api.dropbox.com"
	ContentUploadEndpoint  = "https://content.dropboxapi.com"
)

const ServiceAPIVersion = "2"

type Client struct {
	credentialsPath string
	Credentials     *Credentials
	HTTPClient      HTTPClient
}

func New(opts ...Option) (*Client, error) {
	credentialsPath := ""

	home, err := os.UserHomeDir()
	if err == nil {
		credentialsPath = filepath.Join(home, ".dropbox", "config")
	}

	cli := &Client{
		credentialsPath: credentialsPath,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	cli.parseOptions(opts...)

	credentials, err := loadCredentials(cli.credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %v", err)
	}

	cli.Credentials = credentials

	return cli, nil
}

func (c *Client) CredentialsPath() string {
	return c.credentialsPath
}

func (c *Client) basicAuth() string {
	auth := c.Credentials.AppKey + ":" + c.Credentials.AppSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *Client) bearerAuth() string {
	return "Bearer " + c.Credentials.AccessToken
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
