package domain

import (
	"errors"
	"strings"
	"time"
)

type Tag struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

func NormalizeTagName(name string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return "", errors.New("tag name is empty")
	}
	return n, nil
}

func NormalizeTagNames(names []string) ([]string, error) {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		n, err := NormalizeTagName(name)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out, nil
}
