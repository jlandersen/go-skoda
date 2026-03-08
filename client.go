package skoda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	DefaultBaseURL = "https://mysmob.api.connect.skoda-auto.cz"
	DefaultAPIURL  = DefaultBaseURL + "/api"

	ClientID    = "7f045eee-7003-4379-9968-9355ed2adb06@apps_vw-dilab_com"
	RedirectURI = "myskoda://redirect/login/"

	IdentBaseURL = "https://identity.vwgroup.io"

	AllGenerations = "connectivityGenerations=MOD1&connectivityGenerations=MOD2&connectivityGenerations=MOD3&connectivityGenerations=MOD4"
)

// Client is the main entry point for interacting with the Skoda API.
type Client struct {
	httpClient *http.Client

	baseURL string
	apiURL  string

	mu       sync.Mutex
	tokens   *IDKSession
	email    string
	password string
}

// NewClient creates a new Skoda API client.
//
// Use [Client.Login] or [Client.LoginWithRefreshToken] to authenticate
// before calling any API methods.
func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		apiURL:  DefaultAPIURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// doGet performs an authenticated GET request and decodes the JSON response into dst.
func (c *Client) doGet(ctx context.Context, url string, dst any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return fmt.Errorf("getting access token: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{StatusCode: resp.StatusCode, Body: string(body), URL: url}
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

// APIError is returned when the API responds with a non-200 status code.
type APIError struct {
	StatusCode int
	Body       string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("skoda api: %s returned %d: %s", e.URL, e.StatusCode, e.Body)
}
