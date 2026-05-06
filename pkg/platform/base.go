package platform

import (
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"
)

const DefaultTimeout = 30 * time.Second

// BaseClient holds fields common to all platform clients.
// Each platform client embeds BaseClient and adds platform-specific fields.
type BaseClient struct {
	HTTP    *http.Client
	Log     *slog.Logger
	BaseURL string
	Name    string
}

// NewBaseClient creates a BaseClient with sensible defaults.
func NewBaseClient(name, baseURL string) BaseClient {
	jar, _ := cookiejar.New(nil)
	return BaseClient{
		HTTP: &http.Client{
			Timeout: DefaultTimeout,
			Jar:     jar,
		},
		Log:     slog.Default(),
		BaseURL: baseURL,
		Name:    name,
	}
}

// SetLogger replaces the default logger. Called by the factory during wiring.
func (bc *BaseClient) SetLogger(log *slog.Logger) {
	bc.Log = log
}
