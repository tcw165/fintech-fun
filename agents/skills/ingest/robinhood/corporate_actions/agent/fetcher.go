package agent

import (
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/agent/contract"
	"github.com/tcw165/fintech-fun/agents/skills/ingest/robinhood/corporate_actions/fetch"
)

type LiveFetcher struct{}

func (LiveFetcher) Fetch(url string) (contract.FetchedPage, error) {
	page, err := fetch.FetchTracker(nil, url)
	if err != nil {
		return contract.FetchedPage{}, err
	}
	return contract.FetchedPage{Status: page.Status, URL: page.URL, Bytes: page.Bytes, Text: page.Text}, nil
}

var _ contract.Fetcher = LiveFetcher{}
