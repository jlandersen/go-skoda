package skoda

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// IDKSession holds the JWT tokens from the Skoda identity provider.
type IDKSession struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	IDToken      string `json:"idToken"`
}

// Login authenticates with the Skoda API using email and password.
// This performs the full OIDC PKCE flow against the VW identity provider.
func (c *Client) Login(ctx context.Context, email, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.email = email
	c.password = password

	session, err := c.performLogin(ctx, email, password)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	c.tokens = session
	return nil
}

// LoginWithRefreshToken authenticates using an existing refresh token,
// bypassing the username/password flow.
func (c *Client) LoginWithRefreshToken(ctx context.Context, refreshToken string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := c.refreshTokens(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("refresh token login: %w", err)
	}

	c.tokens = session
	return nil
}

// GetRefreshToken returns the current refresh token, which can be stored
// for later use with [Client.LoginWithRefreshToken].
func (c *Client) GetRefreshToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tokens == nil {
		return "", fmt.Errorf("not authenticated")
	}
	return c.tokens.RefreshToken, nil
}

// accessToken returns a valid access token, refreshing if necessary.
func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tokens == nil {
		return "", fmt.Errorf("not authenticated: call Login or LoginWithRefreshToken first")
	}

	if isTokenExpired(c.tokens.AccessToken) {
		session, err := c.refreshTokens(ctx, c.tokens.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("refreshing token: %w", err)
		}
		c.tokens = session
	}

	return c.tokens.AccessToken, nil
}

func (c *Client) performLogin(ctx context.Context, email, password string) (*IDKSession, error) {
	jar, _ := cookiejar.New(nil)

	// Shared client for the entire OIDC flow — follows redirects and keeps cookies.
	authHTTP := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}

	// Separate no-redirect client for the password step where we follow redirects manually.
	noRedirectHTTP := &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	verifier := generateNonce(43)
	challenge := pkceChallenge(verifier)

	csrf, err := c.oidcAuthorize(ctx, authHTTP, challenge)
	if err != nil {
		return nil, fmt.Errorf("oidc authorize: %w", err)
	}

	csrf, err = c.enterEmail(ctx, authHTTP, csrf, email)
	if err != nil {
		return nil, fmt.Errorf("enter email: %w", err)
	}

	code, err := c.enterPassword(ctx, noRedirectHTTP, csrf, email, password)
	if err != nil {
		return nil, fmt.Errorf("enter password: %w", err)
	}

	return c.exchangeCode(ctx, code, verifier)
}

type csrfState struct {
	csrf       string
	relayState string
	hmac       string
}

// sanitizeURL re-encodes query parameters so that values with spaces are
// properly percent-encoded. Callback URLs extracted from JSON may contain
// raw spaces in values like "scopes=address badge birthdate ...".
func sanitizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.RawQuery = u.Query().Encode()
	return u.String()
}

// The VW identity provider embeds CSRF state in a <script> tag as:
//
//	window._IDK = { csrf_token: "...", templateModel: { hmac: "...", relayState: "..." } }
var (
	csrfRegex       = regexp.MustCompile(`csrf_token:\s*['"]([^'"]+)['"]`)
	relayStateRegex = regexp.MustCompile(`"relayState"\s*:\s*"([^"]+)"`)
	hmacRegex       = regexp.MustCompile(`"hmac"\s*:\s*"([^"]+)"`)
	callbackRegex   = regexp.MustCompile(`"callback"\s*:\s*"([^"]+)"`)
)

func parseCSRF(html string) (csrfState, error) {
	var state csrfState

	m := csrfRegex.FindStringSubmatch(html)
	if m == nil {
		return state, fmt.Errorf("csrf token not found in response")
	}
	state.csrf = m[1]

	m = relayStateRegex.FindStringSubmatch(html)
	if m == nil {
		return state, fmt.Errorf("relay state not found in response")
	}
	state.relayState = m[1]

	m = hmacRegex.FindStringSubmatch(html)
	if m == nil {
		return state, fmt.Errorf("hmac not found in response")
	}
	state.hmac = m[1]

	return state, nil
}

func (c *Client) oidcAuthorize(ctx context.Context, httpClient *http.Client, challenge string) (csrfState, error) {
	params := url.Values{
		"client_id":             {ClientID},
		"nonce":                 {generateNonce(16)},
		"redirect_uri":          {RedirectURI},
		"response_type":         {"code"},
		"scope":                 {"address badge birthdate cars driversLicense dealers email mileage mbb nationalIdentifier openid phone profession profile vin"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"s256"},
		"prompt":                {"login"},
	}

	authorizeURL := IdentBaseURL + "/oidc/v1/authorize?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authorizeURL, nil)
	if err != nil {
		return csrfState{}, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return csrfState{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return csrfState{}, err
	}

	return parseCSRF(string(body))
}

func (c *Client) enterEmail(ctx context.Context, httpClient *http.Client, csrf csrfState, email string) (csrfState, error) {
	form := url.Values{
		"relayState": {csrf.relayState},
		"email":      {email},
		"hmac":       {csrf.hmac},
		"_csrf":      {csrf.csrf},
	}

	loginURL := fmt.Sprintf("%s/signin-service/v1/%s/login/identifier", IdentBaseURL, ClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return csrfState{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return csrfState{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return csrfState{}, err
	}

	return parseCSRF(string(body))
}

func (c *Client) enterPassword(ctx context.Context, httpClient *http.Client, csrf csrfState, email, password string) (string, error) {
	form := url.Values{
		"relayState": {csrf.relayState},
		"email":      {email},
		"password":   {password},
		"hmac":       {csrf.hmac},
		"_csrf":      {csrf.csrf},
	}

	authURL := fmt.Sprintf("%s/signin-service/v1/%s/login/authenticate", IdentBaseURL, ClientID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	for !strings.HasPrefix(location, "myskoda://") {
		if location == "" {
			// Consent pages (marketing, terms) return 200 with a callback URL
			// embedded in the templateModel JSON. Extract it to continue the flow.
			body, _ := io.ReadAll(resp.Body)
			if m := callbackRegex.FindSubmatch(body); m != nil {
				location = sanitizeURL(strings.ReplaceAll(string(m[1]), "&amp;", "&"))
				continue
			}
			return "", fmt.Errorf("no redirect location in auth response (status %d)", resp.StatusCode)
		}
		if strings.Contains(location, "terms-and-conditions") {
			return "", fmt.Errorf("terms and conditions acceptance required: %s", location)
		}

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		if err != nil {
			return "", err
		}

		resp, err = httpClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		location = resp.Header.Get("Location")
	}

	parsed, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parsing redirect URL: %w", err)
	}

	code := parsed.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("no authorization code in redirect URL: %s", location)
	}

	return code, nil
}

func (c *Client) exchangeCode(ctx context.Context, code, verifier string) (*IDKSession, error) {
	data := map[string]string{
		"code":        code,
		"redirectUri": RedirectURI,
		"verifier":    verifier,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	exchangeURL := c.baseURL + "/api/v1/authentication/exchange-authorization-code?tokenType=CONNECT"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, exchangeURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("code exchange failed with status %d: %s", resp.StatusCode, respBody)
	}

	var session IDKSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decoding token response: %w", err)
	}

	return &session, nil
}

func (c *Client) refreshTokens(ctx context.Context, refreshToken string) (*IDKSession, error) {
	data := map[string]string{"token": refreshToken}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	refreshURL := c.baseURL + "/api/v1/authentication/refresh-token?tokenType=CONNECT"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, refreshURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, respBody)
	}

	var session IDKSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("decoding refresh response: %w", err)
	}

	return &session, nil
}

// isTokenExpired checks if a JWT access token is expired or about to expire (within 10 minutes).
// Decodes the payload without verifying signature, same approach as the myskoda Python library.
func isTokenExpired(token string) bool {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return true
	}

	payload := parts[1]
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return true
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return true
	}

	expiry := time.Unix(claims.Exp, 0)
	return time.Now().Add(10 * time.Minute).After(expiry)
}

func generateNonce(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:length]
}

func pkceChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
