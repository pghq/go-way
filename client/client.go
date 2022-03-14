package client

import (
	"context"
	"io"
	"net/http"

	"github.com/pghq/go-tea"
)

const (
	// Version is the version of the client
	Version = "0.0.17"

	// UserAgent is the default user agent for outgoing requests
	UserAgent = "go-way/v" + Version
)

// Get http request
func Get(ctx context.Context, url string) (*http.Response, error) {
	return do(ctx, http.MethodGet, url, nil)
}

// do a http request
func do(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	r, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, tea.Stacktrace(err)
	}

	r.Header.Set("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, tea.Stacktrace(err)
	}

	if resp.StatusCode != 200 {
		return nil, tea.Errf("unexpected refresh response code %d", resp.StatusCode)
	}

	return resp, nil
}
