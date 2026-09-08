// Package cli_params is the corporate_actions command contract. Pure data.
package cli_params

// CliParams is one parsed CLI invocation. Empty Name means default ingest.
type CliParams struct {
	Name   string
	File   string
	DryRun bool
}

// Request is the previous name of CliParams.
type Request = CliParams
