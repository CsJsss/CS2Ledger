package platform

import (
	"context"
	"net/http"

	"github.com/CsJsss/CS2Ledger/pkg/platform/httpclient"
	"github.com/CsJsss/CS2Ledger/pkg/utils/logfx"
)

const DefaultTimeout = httpclient.DefaultTimeout

// BaseClient holds fields common to all platform clients.
// Each platform client embeds BaseClient and adds platform-specific fields.
type BaseClient struct {
	*httpclient.Client
	Log     *logfx.Logger
	BaseURL string
	Name    string
}

// NewBaseClient creates a BaseClient with sensible defaults (cookie jar, 3 retries).
// Extra options are passed to the underlying httpclient.Client.
func NewBaseClient(name, baseURL string, logger *logfx.Logger, opts ...httpclient.Option) BaseClient {
	defaultOpts := make([]httpclient.Option, 0, 5+len(opts))
	defaultOpts = append(defaultOpts,
		httpclient.WithName(name),
		httpclient.WithBaseURL(baseURL),
		httpclient.WithCookieJar(),
		httpclient.WithRetry(httpclient.DefaultMaxRetries),
		httpclient.WithWarnKV(logger.Warn),
	)
	return BaseClient{
		Client:  httpclient.New(append(defaultOpts, opts...)...),
		Log:     logger,
		BaseURL: baseURL,
		Name:    name,
	}
}

// DoRequest delegates to httpclient.Client, syncing BaseURL first
// so that tests can override c.BaseURL after construction.
func (b *BaseClient) DoRequest(ctx context.Context, method, path string, query map[string]string, body []byte, headers http.Header) (int, []byte, error) {
	b.SetBaseURL(b.BaseURL)
	return b.Client.DoRequest(ctx, method, path, query, body, headers)
}
