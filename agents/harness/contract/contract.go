// Package contract defines the agent harness protocol. Pure data and interfaces.
package contract

import "context"

type Request struct {
	Args []string
}

type Result struct {
	Payload any
}

type Skill interface {
	Name() string
	Run(ctx context.Context, req Request) (Result, error)
}

type Runner interface {
	Run(ctx context.Context, skill Skill, req Request) (Result, error)
}

// Manifest is the Agent Skills frontmatter the SDK uses to dispatch tools.
type Manifest struct {
	Name         string
	Description  string
	AllowedTools []string
	DefaultTool  string
	Body         string
}

// Tool is one named action a skill may expose (parse, ingest, fold, …).
type Tool func(ctx context.Context, req Request) (Result, error)

// Toolset is a skill that publishes named tools for the SDK.
type Toolset interface {
	Name() string
	Tools() map[string]Tool
}

// SDK reads a Manifest (from SKILL.md) and runs only allowed tools.
type SDK interface {
	Run(ctx context.Context, manifest Manifest, tools map[string]Tool, req Request) (Result, error)
}
