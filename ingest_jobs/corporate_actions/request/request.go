// Package request is the corporate_actions command contract. Pure data.
package request

// Request is one parsed CLI invocation. Empty Name means default ingest.
type Request struct {
	Name   string
	File   string
	DryRun bool
}
