package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	DefaultTimeout          = 30 * time.Second
	DefaultMaxRetries       = 3
	DefaultRetryInterval    = 500 * time.Millisecond
	DefaultMaxRetryInterval = 30 * time.Second
)

type RetryConfig struct {
	MaxRetries      uint64
	InitialInterval time.Duration
	MaxInterval     time.Duration
}

type Client struct {
	http        *http.Client
	baseURL     string
	name        string
	retryCfg    RetryConfig
	rateLimiter *tokenBucket
	warnKV      func(msg string, keyvals ...any)
}

type Option func(*Client)

func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.http.Timeout = d }
}

func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

func WithCookieJar() Option {
	return func(c *Client) {
		jar, _ := cookiejar.New(nil)
		c.http.Jar = jar
	}
}

func WithRetry(maxRetries uint64) Option {
	return func(c *Client) {
		c.retryCfg.MaxRetries = maxRetries
	}
}

func WithNoRetry() Option {
	return func(c *Client) {
		c.retryCfg.MaxRetries = 0
	}
}

func WithRateLimit(burst int, rps float64) Option {
	return func(c *Client) {
		c.rateLimiter = newTokenBucket(burst, rps)
	}
}

func WithName(name string) Option {
	return func(c *Client) { c.name = name }
}

func WithWarnKV(fn func(msg string, keyvals ...any)) Option {
	return func(c *Client) { c.warnKV = fn }
}

// SetBaseURL allows overriding the base URL after construction (used by BaseClient for test compatibility).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

func New(opts ...Option) *Client {
	c := &Client{
		http: &http.Client{Timeout: DefaultTimeout},
		retryCfg: RetryConfig{
			MaxRetries:      DefaultMaxRetries,
			InitialInterval: DefaultRetryInterval,
			MaxInterval:     DefaultMaxRetryInterval,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// DoRequest executes an HTTP request with optional rate limiting and retry.
func (c *Client) DoRequest(ctx context.Context, method, path string, query map[string]string, body []byte, headers http.Header) (int, []byte, error) {
	if c.rateLimiter != nil {
		if err := c.rateLimiter.acquire(ctx); err != nil {
			return 0, nil, fmt.Errorf("rate limit: %w", err)
		}
	}

	if c.retryCfg.MaxRetries == 0 {
		return c.doRequestOnce(ctx, method, path, query, body, headers)
	}

	var status int
	var respBody []byte
	var netErr error

	op := func() error {
		s, b, err := c.doRequestOnce(ctx, method, path, query, body, headers)
		status = s
		respBody = b
		netErr = err
		if err != nil {
			return err
		}
		if s == 429 || s >= 500 {
			return fmt.Errorf("HTTP %d", s)
		}
		return nil
	}

	eb := backoff.NewExponentialBackOff()
	eb.InitialInterval = c.retryCfg.InitialInterval
	eb.MaxInterval = c.retryCfg.MaxInterval
	eb.MaxElapsedTime = 0
	bo := backoff.WithContext(backoff.WithMaxRetries(eb, c.retryCfg.MaxRetries), ctx)

	_ = backoff.RetryNotify(op, bo, func(err error, d time.Duration) {
		if c.warnKV != nil {
			c.warnKV("request retrying", "name", c.name, "method", method, "path", path, "wait", d, "err", err)
		}
	})

	return status, respBody, netErr
}

func (c *Client) doRequestOnce(ctx context.Context, method, path string, query map[string]string, body []byte, headers http.Header) (int, []byte, error) {
	url := c.baseURL + path
	if len(query) > 0 {
		parts := make([]string, 0, len(query))
		for k, v := range query {
			parts = append(parts, k+"="+v)
		}
		url += "?" + strings.Join(parts, "&")
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return 0, nil, err
	}

	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%s %s %s: %w", c.name, method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}

// --- token bucket rate limiter ---

type tokenBucket struct {
	tokens chan struct{}
	done   chan struct{}
}

func newTokenBucket(burst int, rps float64) *tokenBucket {
	tb := &tokenBucket{
		tokens: make(chan struct{}, burst),
		done:   make(chan struct{}),
	}
	for i := 0; i < burst; i++ {
		tb.tokens <- struct{}{}
	}
	interval := time.Duration(float64(time.Second) / rps)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				select {
				case tb.tokens <- struct{}{}:
				default:
				}
			case <-tb.done:
				return
			}
		}
	}()
	return tb
}

func (tb *tokenBucket) acquire(ctx context.Context) error {
	select {
	case <-tb.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
