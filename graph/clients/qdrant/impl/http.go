package impl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tcw165/fintech-fun/graph/clients/qdrant"
)

// HTTP is the official Qdrant REST client. Child of //graph/clients/qdrant.
type HTTP struct {
	BaseURL string
	Client  *http.Client
}

func NewHTTP(baseURL string) *HTTP {
	return &HTTP{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func URLFromEnv() string {
	if value := os.Getenv("QDRANT_URL"); value != "" {
		return value
	}
	return URL
}

func NewHTTPFromEnv() *HTTP {
	return NewHTTP(URLFromEnv())
}

func (c *HTTP) PutCollection(name string, body map[string]any) (map[string]any, error) {
	return c.do(http.MethodPut, "/collections/"+name, body)
}

func (c *HTTP) Upsert(name string, body map[string]any) (map[string]any, error) {
	return c.do(http.MethodPut, "/collections/"+name+"/points?wait=true", body)
}

func (c *HTTP) do(method, path string, body map[string]any) (map[string]any, error) {
	if c == nil || c.BaseURL == "" {
		return nil, fmt.Errorf("qdrant http client has no base URL")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qdrant %s %s: %s: %s", method, path, resp.Status, payload)
	}
	out := map[string]any{}
	if len(payload) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

var _ qdrant.Client = (*HTTP)(nil)
