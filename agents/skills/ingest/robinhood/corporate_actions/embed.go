package hood_events

import _ "embed"

// SkillMarkdown is the Agent Skills file the harness SDK reads to drive ingest.
//
//go:embed SKILL.md
var SkillMarkdown []byte
