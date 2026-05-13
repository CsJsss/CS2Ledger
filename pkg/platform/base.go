package platform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
	"github.com/cenkalti/backoff/v4"
)

const DefaultTimeout = 30 * time.Second

// BaseClient holds fields common to all platform clients.
// Each platform client embeds BaseClient and adds platform-specific fields.
type BaseClient struct {
	HTTP    *http.Client
	Log     *logfx.Logger
	BaseURL string
	Name    string
}

// NewBaseClient creates a BaseClient with sensible defaults.
func NewBaseClient(name, baseURL string, logger *logfx.Logger) BaseClient {
	jar, _ := cookiejar.New(nil)
	return BaseClient{
		HTTP: &http.Client{
			Timeout: DefaultTimeout,
			Jar:     jar,
		},
		Log:     logger,
		BaseURL: baseURL,
		Name:    name,
	}
}

// DoRequest executes an HTTP request, retrying transient errors (network
// errors, 429 rate limiting, 5xx server errors) with exponential backoff.
func (b *BaseClient) DoRequest(ctx context.Context, method, path string, query map[string]string, body []byte, headers http.Header) (int, []byte, error) {
	var status int
	var respBody []byte
	var netErr error

	op := func() error {
		s, body, err := b.doRequestOnce(ctx, method, path, query, body, headers)
		status = s
		respBody = body
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
	eb.MaxElapsedTime = 0
	bo := backoff.WithContext(backoff.WithMaxRetries(eb, 3), ctx)

	_ = backoff.RetryNotify(op, bo, func(err error, d time.Duration) {
		b.Log.Warn("request retrying", "platform", b.Name, "method", method, "path", path, "wait", d, "err", err)
	})

	return status, respBody, netErr
}

func (b *BaseClient) doRequestOnce(ctx context.Context, method, path string, query map[string]string, body []byte, headers http.Header) (int, []byte, error) {
	url := b.BaseURL + path
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

	resp, err := b.HTTP.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%s %s %s: %w", b.Name, method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, respBody, nil
}
