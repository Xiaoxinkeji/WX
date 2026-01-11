package models

import (
	"errors"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type ArticleDTO struct {
	ID             string
	Title          string
	Content        string
	Status         string
	CreatedAtMs    int64
	UpdatedAtMs    int64
	CurrentVersion int
}

func (a ArticleDTO) ToDomain(tags []domain.Tag) (domain.Article, error) {
	status := domain.ArticleStatus(a.Status)
	if !status.Valid() {
		return domain.Article{}, errors.New("invalid status")
	}

	createdAt := time.UnixMilli(a.CreatedAtMs).UTC()
	updatedAt := time.UnixMilli(a.UpdatedAtMs).UTC()

	return domain.Article{
		ID:             a.ID,
		Title:          a.Title,
		Content:        a.Content,
		Status:         status,
		Tags:           tags,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		CurrentVersion: a.CurrentVersion,
	}, nil
}

type ArticleVersionDTO struct {
	ArticleID   string
	Version     int
	Title       string
	Content     string
	Status      string
	TagsCSV     string
	CreatedAtMs int64
	IsAutoSave  int
}

func (v ArticleVersionDTO) ToDomain() (domain.ArticleVersion, error) {
	status := domain.ArticleStatus(v.Status)
	if !status.Valid() {
		return domain.ArticleVersion{}, errors.New("invalid status")
	}

	tags := parseTagsCSV(v.TagsCSV)
	return domain.ArticleVersion{
		ArticleID:  v.ArticleID,
		Version:    v.Version,
		Title:      v.Title,
		Content:    v.Content,
		Status:     status,
		Tags:       tags,
		CreatedAt:  time.UnixMilli(v.CreatedAtMs).UTC(),
		IsAutoSave: v.IsAutoSave != 0,
	}, nil
}

func parseTagsCSV(csv string) []string {
	if csv == "" {
		return nil
	}
	out := make([]string, 0, 4)
	start := 0
	for i := 0; i <= len(csv); i++ {
		if i == len(csv) || csv[i] == ',' {
			if i > start {
				out = append(out, csv[start:i])
			}
			start = i + 1
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
