// Package cli_params is the corporate_actions command contract. Pure data.
package cli_params

// CliParams is one parsed CLI invocation.
type CliParams struct {
	File   string
	DryRun bool
}
