package skoda

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jlandersen/go-skoda/internal/api"
)

const DefaultBaseURL = "https://public.api.connect.skoda-auto.cz"

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*clientConfig)

// WithBaseURL overrides the public API base URL.
func WithBaseURL(baseURL string) Option {
	return func(config *clientConfig) {
		config.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient overrides the HTTP client used for API requests.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(config *clientConfig) {
		if httpClient != nil {
			config.httpClient = httpClient
		}
	}
}

// Client accesses the MySkoda Public API.
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient creates a public API client using a key generated in the MySkoda app.
func NewClient(apiKey string, options ...Option) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	config := clientConfig{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, option := range options {
		option(&config)
	}

	baseURL, err := url.Parse(config.baseURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q", config.baseURL)
	}

	return &Client{
		httpClient: guardRedirects(config.httpClient),
		baseURL:    config.baseURL,
		apiKey:     apiKey,
	}, nil
}

// guardRedirects copies the HTTP client so the API key cannot follow a redirect
// away from the host it was sent to. net/http strips Authorization and Cookie
// headers across hosts but copies custom headers such as X-API-Key verbatim, so
// the key has to be removed here. The copy leaves a caller-supplied client
// untouched.
func guardRedirects(httpClient *http.Client) *http.Client {
	guarded := *httpClient
	callerCheckRedirect := guarded.CheckRedirect
	guarded.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) > 0 {
			origin := via[0].URL
			if req.URL.Host != origin.Host || req.URL.Scheme != origin.Scheme {
				req.Header.Del("X-API-Key")
			}
		}
		if callerCheckRedirect != nil {
			return callerCheckRedirect(req, via)
		}
		return http.ErrUseLastResponse
	}
	return &guarded
}

func (c *Client) doGet(ctx context.Context, requestURL string, dst any) (ResponseMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return ResponseMetadata{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json, application/problem+json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ResponseMetadata{}, fmt.Errorf("sending request: %w", err)
	}

	metadata := responseMetadata(resp.Header)
	body, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return metadata, fmt.Errorf("reading response: %w", readErr)
	}
	if closeErr != nil {
		return metadata, fmt.Errorf("closing response: %w", closeErr)
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
			URL:        requestURL,
			Metadata:   metadata,
		}
		var problem ProblemDetail
		if err := json.Unmarshal(body, &problem); err == nil {
			apiErr.Problem = &problem
		}
		return metadata, apiErr
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return metadata, fmt.Errorf("decoding response: %w", err)
	}
	return metadata, nil
}

// ResponseMetadata contains API-key and rate-limit response headers.
type ResponseMetadata struct {
	APIKeyExpiresAt    string `json:"-"`
	RateLimitLimit     string `json:"-"`
	RateLimitRemaining string `json:"-"`
	RateLimitReset     string `json:"-"`
	RetryAfter         string `json:"-"`
}

func responseMetadata(header http.Header) ResponseMetadata {
	return ResponseMetadata{
		APIKeyExpiresAt:    header.Get("X-API-Key-Expires-At"),
		RateLimitLimit:     header.Get("RateLimit-Limit"),
		RateLimitRemaining: header.Get("RateLimit-Remaining"),
		RateLimitReset:     header.Get("RateLimit-Reset"),
		RetryAfter:         header.Get("Retry-After"),
	}
}

// ProblemDetail is an RFC 9457 error response from the public API.
type ProblemDetail = api.ProblemDetail

// APIError is returned when the public API responds with a non-200 status code.
type APIError struct {
	StatusCode int
	Body       string
	URL        string
	Problem    *ProblemDetail
	Metadata   ResponseMetadata
}

func (e *APIError) Error() string {
	message := ""
	if e.Problem != nil {
		if e.Problem.Detail != nil {
			message = *e.Problem.Detail
		} else if e.Problem.Title != nil {
			message = *e.Problem.Title
		}
	}
	if message == "" {
		message = e.Body
	}
	return fmt.Sprintf("skoda public API: %s returned %d: %s", e.URL, e.StatusCode, message)
}
