// Copyright 2026 Raphael Seebacher
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ltosapi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/raphaelthomas/meinberg_ltos_exporter/pkg/ltosapi/models"
)

const (
	apiStatusPath         = "/api/status"
	maxStatusResponseSize = 1 << 20 // 1 MiB
)

// Client represents a Meinberg LTOS API client
type Client struct {
	baseURL       url.URL
	authBasicUser string
	authBasicPass string
	httpClient    *http.Client
}

// Target returns the target base URL of the Meinberg LTOS API client
func (c *Client) Target() string {
	return c.baseURL.String()
}

// NewClient creates a new Meinberg LTOS API client
func NewClient(baseURL string, authBasicUser, authBasicPass string, ignoreSSLVerify bool) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: ignoreSSLVerify}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("invalid base URL: must include URL scheme and host")
	}
	if parsedURL.User != nil {
		return nil, errors.New("invalid base URL: must not contain credentials, use --auth-user and --auth-pass instead")
	}

	return &Client{
		baseURL:       *parsedURL,
		authBasicUser: authBasicUser,
		authBasicPass: authBasicPass,
		httpClient: &http.Client{
			Transport: transport,
		},
	}, nil
}

// FetchStatus fetches the target status from the Meinberg LTOS API
func (c *Client) FetchStatus(ctx context.Context, logger *slog.Logger) (*models.StatusResponse, error) {
	statusURL := c.baseURL.JoinPath(apiStatusPath).String()
	logger = logger.With("url", statusURL)

	logger.Debug("Fetching status from Meinberg LTOS device API")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, err
	}

	if c.authBasicUser != "" && c.authBasicPass != "" {
		req.SetBasicAuth(c.authBasicUser, c.authBasicPass)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Warn("Failed to close response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logger.Warn("Unexpected status code from Meinberg LTOS device API", "status_code", resp.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read one byte past the limit to tell an oversized response from one
	// that is exactly at the limit.
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxStatusResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read status response: %w", err)
	}
	if len(body) > maxStatusResponseSize {
		return nil, fmt.Errorf("status response exceeds %d bytes", maxStatusResponseSize)
	}

	var data models.StatusResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status response: %w", err)
	}

	logger.Debug("Successfully fetched status from Meinberg LTOS device API")
	return &data, nil
}
