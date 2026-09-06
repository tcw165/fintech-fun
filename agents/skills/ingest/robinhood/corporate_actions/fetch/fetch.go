// Package fetch downloads the Robinhood corporate-actions tracker. Child of hood_events.
package fetch

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions"
)

const maxAttempts = 3

type Result struct {
	Status int    `json:"status_code"`
	URL    string `json:"url"`
	Bytes  int    `json:"bytes"`
	Text   string `json:"text"`
}

func FetchTracker(client *http.Client, url string) (Result, error) {
	if url == "" {
		url = hood_events.TrackerURL
	}
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	var last error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := getOnce(client, url)
		if err == nil {
			return result, nil
		}
		last = err
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}
	return Result{}, last
}

func getOnce(client *http.Client, url string) (Result, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("User-Agent", "fintech-fun-corporate-actions/0.1")
	req.Header.Set("Accept", "text/html,text/plain;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return Result{}, err
	}
	if resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("fetch %s: %s", url, resp.Status)
	}
	text := string(body)
	if !strings.Contains(strings.ToLower(text), "corporate") && !strings.Contains(text, "September") {
		return Result{}, fmt.Errorf("fetch %s: page did not look like the tracker", url)
	}
	return Result{Status: resp.StatusCode, URL: url, Bytes: len(body), Text: text}, nil
}
