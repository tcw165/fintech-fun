// Package skillmd parses Agent Skills SKILL.md files. Pure data; no runner.
//
// Format: YAML frontmatter (name + description required) then a Markdown body.
// See https://agentskills.io/specification
package skillmd

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	maxName        = 64
	maxDescription = 1024
)

// Document is a parsed SKILL.md. Frontmatter fields the spec does not
// recognize are ignored.
type Document struct {
	Name          string
	Description   string
	License       string
	Compatibility string
	AllowedTools  []string
	Metadata      map[string]string
	Body          string
}

// DefaultTool is metadata["default-tool"], else the first allowed tool.
func (d Document) DefaultTool() string {
	if d.Metadata != nil {
		if name := strings.TrimSpace(d.Metadata["default-tool"]); name != "" {
			return name
		}
	}
	if len(d.AllowedTools) > 0 {
		return d.AllowedTools[0]
	}
	return ""
}

// Allows reports whether tool is listed in allowed-tools. An empty list
// allows every registered tool (the skill has not restricted the set).
func (d Document) Allows(tool string) bool {
	if len(d.AllowedTools) == 0 {
		return true
	}
	for _, name := range d.AllowedTools {
		if name == tool {
			return true
		}
	}
	return false
}

// Parse reads a SKILL.md byte slice. It accepts the Agent Skills subset:
// scalar frontmatter keys and a string-to-string metadata map.
func Parse(raw []byte) (Document, error) {
	text := strings.TrimPrefix(string(raw), "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	trimmed := strings.TrimLeft(text, " \t\n")
	if !strings.HasPrefix(trimmed, "---") {
		return Document{}, fmt.Errorf("skillmd: missing YAML frontmatter")
	}
	rest := strings.TrimPrefix(trimmed, "---")
	if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return Document{}, fmt.Errorf("skillmd: unclosed YAML frontmatter")
	}
	front := rest[:end]
	body := strings.TrimPrefix(rest[end+len("\n---"):], "\n")
	body = strings.TrimRight(body, "\n") + "\n"

	fields, err := parseFrontmatter(front)
	if err != nil {
		return Document{}, err
	}
	doc := Document{
		Name:          fields.scalars["name"],
		Description:   fields.scalars["description"],
		License:       fields.scalars["license"],
		Compatibility: fields.scalars["compatibility"],
		AllowedTools:  splitTools(fields.scalars["allowed-tools"]),
		Metadata:      fields.metadata,
		Body:          body,
	}
	if err := Validate(doc); err != nil {
		return Document{}, err
	}
	return doc, nil
}

// Validate checks Agent Skills name and description constraints.
func Validate(doc Document) error {
	if err := validateName(doc.Name); err != nil {
		return err
	}
	if doc.Description == "" {
		return fmt.Errorf("skillmd: description is required")
	}
	if len(doc.Description) > maxDescription {
		return fmt.Errorf("skillmd: description exceeds %d characters", maxDescription)
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("skillmd: name is required")
	}
	if len(name) > maxName {
		return fmt.Errorf("skillmd: name exceeds %d characters", maxName)
	}
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		return fmt.Errorf("skillmd: name must not start or end with a hyphen")
	}
	if strings.Contains(name, "--") {
		return fmt.Errorf("skillmd: name must not contain consecutive hyphens")
	}
	for _, r := range name {
		if r == '-' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		return fmt.Errorf("skillmd: name %q must be lowercase alphanumeric with hyphens", name)
	}
	return nil
}

type frontmatter struct {
	scalars  map[string]string
	metadata map[string]string
}

func parseFrontmatter(front string) (frontmatter, error) {
	out := frontmatter{scalars: map[string]string{}, metadata: map[string]string{}}
	lines := strings.Split(front, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		if line == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if indent(line) != 0 {
			return frontmatter{}, fmt.Errorf("skillmd: unexpected indent %q", line)
		}
		key, value, ok := splitKV(line)
		if !ok {
			return frontmatter{}, fmt.Errorf("skillmd: invalid frontmatter line %q", line)
		}
		if key == "metadata" {
			if value != "" {
				return frontmatter{}, fmt.Errorf("skillmd: metadata must be a mapping")
			}
			meta, next, err := parseMapping(lines, i+1)
			if err != nil {
				return frontmatter{}, err
			}
			out.metadata = meta
			i = next - 1
			continue
		}
		out.scalars[key] = unquote(value)
	}
	return out, nil
}

func parseMapping(lines []string, start int) (map[string]string, int, error) {
	meta := map[string]string{}
	i := start
	for ; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		if line == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		if indent(line) == 0 {
			break
		}
		key, value, ok := splitKV(strings.TrimSpace(line))
		if !ok || value == "" && lookIndentedChild(lines, i+1) {
			return nil, 0, fmt.Errorf("skillmd: nested metadata is not supported")
		}
		if !ok {
			return nil, 0, fmt.Errorf("skillmd: invalid metadata line %q", line)
		}
		meta[key] = unquote(value)
	}
	return meta, i, nil
}

func lookIndentedChild(lines []string, start int) bool {
	for i := start; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], " \t")
		if line == "" {
			continue
		}
		return indent(line) > 0
	}
	return false
}

func splitKV(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", false
	}
	for _, r := range key {
		if r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", "", false
		}
	}
	return key, strings.TrimSpace(line[idx+1:]), true
}

func indent(line string) int {
	n := 0
	for _, r := range line {
		if r == ' ' {
			n++
			continue
		}
		if r == '\t' {
			n += 2
			continue
		}
		break
	}
	return n
}

func unquote(value string) string {
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func splitTools(value string) []string {
	if value == "" {
		return nil
	}
	var tools []string
	for _, part := range strings.Fields(value) {
		if part != "" {
			tools = append(tools, part)
		}
	}
	return tools
}
