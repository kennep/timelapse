package client

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to the timelapse server",
	Long:  `Log in to the timelapse server using the specified login provider`,
	Args:  cobra.NoArgs,
	Run: runCommand(func(cmd *cobra.Command, args []string) error {
		return login()
	}),
}

const authURL = "https://accounts.google.com/o/oauth2/v2/auth"
const tokenURL = "https://www.googleapis.com/oauth2/v4/token"

// Client ID for the timelapse command-line client
const clientID = "91541969634-fih4ctg9nblm8qb9qdpn55eenlketoul.apps.googleusercontent.com"

// Note: the client "secret" in this case is considered public knowledge, therefore it
// is embedded in the source code.
const clientSecret = "1rDnwiqL9ci2z0VUatH2Cxzr"

const codeChallengeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._"

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

var codeChallengeEncoding *base64.Encoding

func generateCodeChallenge() string {
	c := 40
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return codeChallengeEncoding.EncodeToString(b)
}

func initCodeChallengeEncoding() {
	codeChallengeEncoding = base64.NewEncoding(codeChallengeAlphabet).WithPadding(base64.NoPadding)
}

func writeError(w http.ResponseWriter, message string) {
	fmt.Fprintf(w, "<html><head></head><body><p>%s</p><p><a href=\"login\">Try again</a></p></body></html>\n", message)
}

func refreshTokens(refreshToken string) error {
	tokenParams := make(url.Values)
	tokenParams.Add("client_id", clientID)
	tokenParams.Add("client_secret", clientSecret)
	tokenParams.Add("grant_type", "refresh_token")
	tokenParams.Add("refresh_token", refreshToken)

	resp, err := http.PostForm(tokenURL, tokenParams)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New("Refresh token failed " + string(body))
	}

	var tokens tokenResponse

	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return err
	}

	credentials, err := GetCredentials()
	if err != nil {
		return err
	}

	if tokens.RefreshToken == "" {
		tokens.RefreshToken = refreshToken
	}

	credentials.SetProviderCredentialsFromToken("google", &tokens)
	err = credentials.Store()
	if err != nil {
		return err
	}
	return nil
}

func login() error {
	initCodeChallengeEncoding()

	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return err
	}

	resultChan := make(chan error)

	var server http.Server
	var codeChallenge string
	redirectURI := fmt.Sprintf("http://%s/post_login", listener.Addr().String())

	servemux := http.NewServeMux()
	server.Handler = servemux

	servemux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "<html><head></head><body><a href=\"login\">Log in</a></body></html>\n")
	})

	servemux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		codeChallenge = generateCodeChallenge()

		params := make(url.Values)
		params.Add("client_id", clientID)
		params.Add("redirect_uri", redirectURI)
		params.Add("response_type", "code")
		params.Add("scope", "email")
		params.Add("state", "n="+codeChallenge)
		params.Add("code_challenge_method", "plain")
		params.Add("code_challenge", codeChallenge)

		url := authURL + "?" + params.Encode()
		http.Redirect(w, r, url, 302)
	})

	servemux.HandleFunc("/post_login", func(w http.ResponseWriter, r *http.Request) {
		params, _ := url.ParseQuery(r.URL.RawQuery)

		tokenParams := make(url.Values)
		tokenParams.Add("code", params.Get("code"))
		tokenParams.Add("client_id", clientID)
		tokenParams.Add("client_secret", clientSecret)
		tokenParams.Add("redirect_uri", redirectURI)
		tokenParams.Add("grant_type", "authorization_code")
		tokenParams.Add("code_verifier", codeChallenge)

		resp, err := http.PostForm(tokenURL, tokenParams)
		if err != nil {
			writeError(w, "There was an error while logging in.")
			resultChan <- err
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			writeError(w, "The login provider did not send a complete response.")
			resultChan <- err
			return
		}

		if resp.StatusCode != 200 {
			errorMsg := "Getting access token failed " + string(body)
			writeError(w, errorMsg)
			resultChan <- errors.New(errorMsg)
		}

		var tokens tokenResponse

		err = json.Unmarshal(body, &tokens)
		if err != nil {
			writeError(w, "The login provider sent a malformed response.")
			resultChan <- err
			return
		}

		credentials, err := GetCredentials()
		if err != nil {
			writeError(w, "Login was successful, but could not save access tokens.")
			resultChan <- err
			return
		}

		credentials.SetProviderCredentialsFromToken("google", &tokens)
		err = credentials.Store()
		if err != nil {
			writeError(w, "Login was successful, but could not save access tokens.")
			resultChan <- err
			return
		}

		fmt.Fprintf(w, "<html><head></head><body>You have successfully logged in to Timelapse. This browser window can now be closed.</body></html>\n")

		resultChan <- nil
	})

	fmt.Fprintf(os.Stderr, "Please open http://%s/login in a web browser log in.\n", listener.Addr().String())

	go func() {
		err = server.Serve(listener)
		if err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Login error: %s\n", err)
		}
	}()
	err = <-resultChan
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login error: %s\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "Login was successful.\n")
	}

	return nil
}
