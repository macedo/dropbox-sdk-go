package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"github.com/cli/browser"
	"github.com/macedo/dropbox-sdk-go/dropbox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

var authorizeCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate user through oauth2 flow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return authenticateAction(os.Stdout)
	},
}

func authenticateAction(w io.Writer) error {
	codeCh := make(chan string)
	errCh := make(chan error)
	exitCh := make(chan bool)

	redirectURI := "http://localhost:" + viper.GetString("port")

	startServer(context.Background(), codeCh, errCh, exitCh)

	fmt.Fprintf(w, "Press any key to open the browser to login or 'q' to exit:\n")

	input, err := readInput()
	if err != nil {
		return fmt.Errorf("failed to read input: %v", err)
	}

	if input == "q" || input == "Q" {
		return errors.New("exit")
	}

	authorizeURL, _ := url.Parse("https://www.dropbox.com/oauth2/authorize")
	query := authorizeURL.Query()
	query.Add("client_id", viper.GetString("client-id"))
	query.Add("redirect_uri", redirectURI)
	query.Add("response_type", "code")
	query.Add("token_access_type", "offline")
	authorizeURL.RawQuery = query.Encode()

	fmt.Fprintf(w, "Opening browser to %s\n", authorizeURL)
	if err := browser.OpenURL(authorizeURL.String()); err != nil {
		return fmt.Errorf("failed to open browser for authorization: %v", err)
	}

	fmt.Fprintln(w, "Waiting fot authorization... ")

	select {
	case code := <-codeCh:
		cli := dropbox.New()
		out, err := cli.OAuth2Token(&dropbox.OAuth2TokenInput{
			ClientID:     viper.GetString("client-id"),
			ClientSecret: viper.GetString("client-secret"),
			Code:         code,
			GrantType:    "authorization_code",
			RedirectURI:  redirectURI,
		})
		if err != nil {
			return fmt.Errorf("failed to get acess token from dropbox: %v", err)
		}

		f, err := os.Create(viper.ConfigFileUsed())
		if err != nil {
			return nil
		}

		if err := writeCredentials(f, &Credentials{
			ClientID:     viper.GetString("client-id"),
			ClientSecret: viper.GetString("client-secret"),
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
		}); err != nil {
			return err
		}

		fmt.Fprintln(w, "Authenticated!!!")
		return nil

	case err := <-errCh:
		return err

	case <-exitCh:
		return nil
	}
}

func readInput() (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	b := make([]byte, 1)
	_, err = os.Stdin.Read(b)
	if err != nil {
		return "", err
	}

	return string(b[0]), nil
}

func startServer(ctx context.Context, codeCh chan<- string, errCh chan<- error, exitCh chan<- bool) {
	server := &http.Server{
		Addr: ":" + viper.GetString("port"),
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if code, ok := r.URL.Query()["code"]; ok {
			codeCh <- code[0]
		}

		io.WriteString(w, "You can close this page and return to your CLI. It now be authenticated")
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to start server: %v", err)
		}
	}()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			errCh <- fmt.Errorf("failed to shuwdown server: %v", err)
		}

		exitCh <- true
	}()
}

type Credentials struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
}

func writeCredentials(out io.Writer, c *Credentials) error {
	credentialsTmpl := `dropbox-client-id="{{.ClientID}}"
dropbox-client-secret="{{.ClientSecret}}"
dropbox-access-token="{{.AccessToken}}"
dropbox-refresh-token="{{.RefreshToken}}"
`
	tmpl, _ := template.New("").Parse(credentialsTmpl)

	return tmpl.Execute(out, c)
}

func init() {
	browser.Stdout = nil
	browser.Stderr = nil

	authorizeCmd.PersistentFlags().String("client-id", "", "dropbox application client id")
	viper.BindPFlag("client-id", authorizeCmd.PersistentFlags().Lookup("client-id"))

	authorizeCmd.PersistentFlags().String("client-secret", "", "dropbox application client secret")
	viper.BindPFlag("client-secret", authorizeCmd.PersistentFlags().Lookup("client-secret"))

	authorizeCmd.PersistentFlags().String("port", "8080", "listening port for redirect uri server")
	viper.BindPFlag("port", authorizeCmd.PersistentFlags().Lookup("port"))

	rootCmd.AddCommand(authorizeCmd)
}
