// Package cli_params is the corporate_actions command contract. Pure data.
package cli_params

// Request is one parsed CLI invocation. Empty Name means default ingest.
type Request struct {
	Name   string
	File   string
	DryRun bool
}
