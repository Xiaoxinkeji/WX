package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	MaxTitleLength   = 500
	MaxContentLength = 1000000 // 1MB
	MaxTagCount      = 50
	MaxTagLength     = 100
)

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
)

type Article struct {
	ID            string
	Title         string
	Content       string
	Status        ArticleStatus
	Tags          []Tag
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CurrentVersion int
}

type ArticleVersion struct {
	ArticleID  string
	Version    int
	Title      string
	Content    string
	Status     ArticleStatus
	Tags       []string
	CreatedAt  time.Time
	IsAutoSave bool
}

func (s ArticleStatus) Valid() bool {
	switch s {
	case ArticleStatusDraft, ArticleStatusPublished:
		return true
	default:
		return false
	}
}

func ValidateArticleFields(status ArticleStatus, title, content string) error {
	if !status.Valid() {
		return errors.New("invalid status")
	}

	// Validate length limits
	if len(title) > MaxTitleLength {
		return fmt.Errorf("title too long: max %d characters, got %d", MaxTitleLength, len(title))
	}

	if len(content) > MaxContentLength {
		return fmt.Errorf("content too long: max %d characters, got %d", MaxContentLength, len(content))
	}

	if status == ArticleStatusPublished {
		if strings.TrimSpace(title) == "" {
			return errors.New("title required for published article")
		}
		if strings.TrimSpace(content) == "" {
			return errors.New("content required for published article")
		}
	}

	return nil
}
