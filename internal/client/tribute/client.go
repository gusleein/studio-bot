package tribute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://tribute.tg/api/v1"

const (
	headerAPIKey      = "Api-Key"
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

// Client — HTTP-клиент Tribute Shop API.
type Client struct {
	baseURL    string
	apiKey     string
	shopID     uint64
	httpClient *http.Client
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

func WithShopID(shopID uint64) Option {
	return func(c *Client) {
		c.shopID = shopID
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{Timeout: d}
			return
		}
		cloned := *c.httpClient
		cloned.Timeout = d
		c.httpClient = &cloned
	}
}

func New(apiKey string, opts ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrEmptyAPIKey
	}

	c := &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return c, nil
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, reqBody any, out any) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("tribute: url: %w", err)
	}

	q := u.Query()
	for k, vs := range query {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	if c.shopID != 0 && q.Get("shopId") == "" {
		q.Set("shopId", fmt.Sprintf("%d", c.shopID))
	}
	u.RawQuery = q.Encode()

	var body io.Reader
	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("tribute: marshal request: %w", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return fmt.Errorf("tribute: new request: %w", err)
	}
	req.Header.Set(headerAPIKey, c.apiKey)
	if reqBody != nil {
		req.Header.Set(headerContentType, contentTypeJSON)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tribute: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("tribute: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("tribute: unmarshal response: %w", err)
	}
	return nil
}

func parseAPIError(status int, body []byte) error {
	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err == nil && (parsed.Error != "" || parsed.Message != "") {
		return &APIError{StatusCode: status, Code: parsed.Error, Message: parsed.Message}
	}
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = http.StatusText(status)
	}
	return &APIError{StatusCode: status, Message: msg}
}
