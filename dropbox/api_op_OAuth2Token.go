package dropbox

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/schema"
)

func (c *Client) OAuth2Token(params *OAuth2TokenInput) (*OAuth2TokenOutput, error) {
	var out OAuth2TokenOutput

	if params == nil {
		params = &OAuth2TokenInput{}
	}

	body := url.Values{}
	if err := schema.NewEncoder().Encode(params, body); err != nil {
		return nil, err
	}

	tokenURL, _ := url.Parse(c.AuthenticationEndpoint)
	tokenURL.Path = "/oauth2/token"

	request, err := http.NewRequest(
		http.MethodPost,
		tokenURL.String(),
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return nil, err
	}

	err = c.sendRequest(request, &out)

	return &out, err
}

type OAuth2TokenInput struct {
	ClientID          string `schema:"client_id,omitempty"`
	ClientSecret      string `schema:"client_secret,omitempty"`
	Code              string `schema:"code,omitempty"`
	CodeChallenge     string `schema:"code_challenge,omitempty"`
	CodeVerifier      string `schema:"code_verifier,omitempty"`
	ExpirationMinutes string `schema:"expiration_minutes,omitempty"`
	GrantType         string `schema:"grant_type,omitempty"`
	RedirectURI       string `schema:"redirect_uri,omitempty"`
	RefreshToken      string `schema:"refresh_token,omitempty"`
	Scope             string `schema:"scope,omitempty"`
}

type OAuth2TokenOutput struct {
	AccountID    string `json:"account_id"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	UID          string `json:"uid"`
}
