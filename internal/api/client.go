package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// ConfigDir returns the path to ~/.xseek
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".xseek")
}

// ConfigPath returns the path to ~/.xseek/config
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config")
}

// readConfigKey reads a key=value from ~/.xseek/config
func readConfigKey(key string) string {
	f, err := os.Open(ConfigPath())
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// SaveConfig writes a key=value to ~/.xseek/config (creates or updates)
func SaveConfig(key, value string) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create %s: %w", dir, err)
	}

	configPath := ConfigPath()
	lines := []string{}
	found := false

	// Read existing config
	if f, err := os.Open(configPath); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
				lines = append(lines, key+"="+value)
				found = true
			} else {
				lines = append(lines, line)
			}
		}
		f.Close()
	}

	if !found {
		lines = append(lines, key+"="+value)
	}

	return os.WriteFile(configPath, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

func NewClient() (*Client, error) {
	// Priority: env var > ~/.xseek/config
	apiKey := os.Getenv(EnvAPIKey)
	if apiKey == "" {
		apiKey = readConfigKey("api_key")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no API key found\n\nAuthenticate with:\n  xseek login YOUR_API_KEY\n\nOr set the environment variable:\n  export %s=your_api_key\n\nGet a key at: https://www.xseek.io/dashboard/api-keys", EnvAPIKey)
	}

	baseURL := os.Getenv(EnvBaseURL)
	if baseURL == "" {
		baseURL = readConfigKey("api_url")
	}
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

func (c *Client) PostJSON(path string, body interface{}, v interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "xseek-cli/"+Version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 {
		return fmt.Errorf("unauthorized — check your XSEEK_API_KEY")
	}
	if resp.StatusCode == 403 {
		return fmt.Errorf("forbidden — your API key may not have the required privileges")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return json.Unmarshal(respBody, v)
}

var Version = "dev"
