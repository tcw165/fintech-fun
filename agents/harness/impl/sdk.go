package impl

import (
	"context"
	"fmt"

	"github.com/tcw165/fintech-fun/agents/harness/contract"
	"github.com/tcw165/fintech-fun/agents/harness/skillmd"
)

type SDK struct{}

func NewSDK() *SDK {
	return &SDK{}
}

// ManifestFrom copies a parsed SKILL.md into the harness Manifest the SDK runs.
func ManifestFrom(doc skillmd.Document) contract.Manifest {
	return contract.Manifest{
		Name:         doc.Name,
		Description:  doc.Description,
		AllowedTools: append([]string(nil), doc.AllowedTools...),
		DefaultTool:  doc.DefaultTool(),
		Body:         doc.Body,
	}
}

func (s *SDK) Run(ctx context.Context, manifest contract.Manifest, tools map[string]contract.Tool, req contract.Request) (contract.Result, error) {
	name := manifest.DefaultTool
	args := req.Args
	if len(req.Args) > 0 {
		name = req.Args[0]
	} else if name != "" {
		args = []string{name}
	}
	if name == "" {
		return contract.Result{}, fmt.Errorf("skill %s: no tool specified", manifest.Name)
	}
	if !allows(manifest, name) {
		return contract.Result{}, fmt.Errorf("skill %s: tool %q is not allowed", manifest.Name, name)
	}
	tool, ok := tools[name]
	if !ok {
		return contract.Result{}, fmt.Errorf("skill %s: unknown tool %q", manifest.Name, name)
	}
	return tool(ctx, contract.Request{Args: args})
}

func allows(manifest contract.Manifest, name string) bool {
	if len(manifest.AllowedTools) == 0 {
		return true
	}
	for _, tool := range manifest.AllowedTools {
		if tool == name {
			return true
		}
	}
	return false
}

var _ contract.SDK = (*SDK)(nil)
