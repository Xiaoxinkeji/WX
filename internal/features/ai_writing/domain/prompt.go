package domain

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	Role    Role
	Content string
}

type PromptMessage struct {
	Role     Role
	Template string
}

type Prompt struct {
	Name     string
	Messages []PromptMessage
}

var placeholderRE = regexp.MustCompile(`{{\s*([a-zA-Z0-9_]+)\s*}}`)

func (p Prompt) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("prompt name is required")
	}
	if len(p.Messages) == 0 {
		return errors.New("prompt messages are required")
	}
	for i, m := range p.Messages {
		if m.Role == "" {
			return fmt.Errorf("prompt message %d: role is required", i)
		}
		if strings.TrimSpace(m.Template) == "" {
			return fmt.Errorf("prompt message %d: template is required", i)
		}
	}
	return nil
}

func (p Prompt) RequiredVariables() []string {
	set := make(map[string]struct{})
	for _, m := range p.Messages {
		for _, match := range placeholderRE.FindAllStringSubmatch(m.Template, -1) {
			if len(match) == 2 {
				set[match[1]] = struct{}{}
			}
		}
	}
	vars := make([]string, 0, len(set))
	for k := range set {
		vars = append(vars, k)
	}
	sort.Strings(vars)
	return vars
}

func (p Prompt) Render(vars map[string]string) ([]Message, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	out := make([]Message, 0, len(p.Messages))
	for i, m := range p.Messages {
		rendered, err := RenderTemplate(m.Template, vars)
		if err != nil {
			return nil, fmt.Errorf("prompt message %d: %w", i, err)
		}
		out = append(out, Message{Role: m.Role, Content: rendered})
	}
	return out, nil
}

func RenderTemplate(template string, vars map[string]string) (string, error) {
	if template == "" {
		return "", errors.New("template is empty")
	}
	if vars == nil {
		vars = map[string]string{}
	}

	missing := make(map[string]struct{})
	out := placeholderRE.ReplaceAllStringFunc(template, func(token string) string {
		match := placeholderRE.FindStringSubmatch(token)
		if len(match) != 2 {
			return token
		}
		key := match[1]
		val, ok := vars[key]
		if !ok {
			missing[key] = struct{}{}
			return token
		}
		return val
	})
	if len(missing) > 0 {
		keys := make([]string, 0, len(missing))
		for k := range missing {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return "", fmt.Errorf("missing variables: %s", strings.Join(keys, ","))
	}
	return out, nil
}
