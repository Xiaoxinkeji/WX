package domain

import (
	"errors"
	"fmt"
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
	if len(n) > MaxTagLength {
		return "", fmt.Errorf("tag name too long: max %d characters, got %d", MaxTagLength, len(n))
	}
	return n, nil
}

func NormalizeTagNames(names []string) ([]string, error) {
	if len(names) > MaxTagCount {
		return nil, fmt.Errorf("too many tags: max %d, got %d", MaxTagCount, len(names))
	}

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
