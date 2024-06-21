package dropbox

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Resolver interface {
	load(*Credentials) error
	setNext(Resolver) Resolver
}

type Credentials struct {
	AppKey       string `toml:"dropbox-app-key"`
	AppSecret    string `toml:"dropbox-app-secret"`
	AccessToken  string `toml:"dropbox-access-token"`
	RefreshToken string `toml:"dropox-refresh-token"`
}

type FileResolver struct {
	next Resolver
	path string
}

func (r *FileResolver) load(c *Credentials) error {
	f, err := os.ReadFile(r.path)
	if err != nil {
		return fmt.Errorf("failed to load credentials file %q: %v", r.path, err)
	}

	if err := toml.Unmarshal(f, c); err != nil {
		return fmt.Errorf("failed to parse credentials file %q: %v", r.path, err)
	}

	if r.next != nil {
		return r.next.load(c)
	}

	return nil
}

func (r *FileResolver) setNext(resolver Resolver) Resolver {
	r.next = resolver
	return resolver
}

type EnvResolver struct {
	next Resolver
}

func (r *EnvResolver) load(c *Credentials) error {
	var k string

	if k = os.Getenv("DROPBOX_APP_KEY"); k != "" {
		c.AppKey = k
	}

	if k = os.Getenv("DROPBOX_APP_SECRET"); k != "" {
		c.AppSecret = k
	}

	if k = os.Getenv("DROPBOX_ACCESS_TOKEN"); k != "" {
		c.AccessToken = k
	}

	if k = os.Getenv("DROPBOX_REFRESH_TOKEN"); k != "" {
		c.RefreshToken = k
	}

	if r.next != nil {
		return r.load(c)
	}

	return nil
}

func (r *EnvResolver) setNext(resolver Resolver) Resolver {
	r.next = resolver
	return resolver
}

type CredentialsResolver struct {
	next Resolver
}

func (r *CredentialsResolver) load(c *Credentials) error {
	fmt.Println("credentials resolver")
	if r.next != nil {
		return r.next.load(c)
	}

	return nil
}

func (r *CredentialsResolver) setNext(resolver Resolver) Resolver {
	r.next = resolver
	return resolver
}

func loadCredentials(path string) (*Credentials, error) {
	var credentials Credentials

	fileResolver := &FileResolver{path: path}
	envResolver := &EnvResolver{}

	fileResolver.setNext(envResolver)
	err := fileResolver.load(&credentials)

	return &credentials, err
}
