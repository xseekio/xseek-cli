package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	DefaultBaseURL = "https://www.xseek.io/api/v1"
	EnvAPIKey      = "XSEEK_API_KEY"
	EnvBaseURL     = "XSEEK_API_URL"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv(EnvAPIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("missing %s environment variable\n\nSet it with:\n  export %s=your_api_key\n\nGet a key at: https://www.xseek.io/dashboard/api-keys", EnvAPIKey, EnvAPIKey)
	}

	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (c *Client) Get(path string, params map[string]string) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "xseek-cli/"+Version)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("unauthorized — check your XSEEK_API_KEY")
	}
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("forbidden — your API key may not have the required privileges")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (c *Client) GetJSON(path string, params map[string]string, v interface{}) error {
	body, err := c.Get(path, params)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

var Version = "dev"
