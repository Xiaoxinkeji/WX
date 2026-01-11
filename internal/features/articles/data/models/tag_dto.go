package models

import (
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type TagDTO struct {
	ID          string
	Name        string
	CreatedAtMs int64
}

func (t TagDTO) ToDomain() domain.Tag {
	return domain.Tag{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: time.UnixMilli(t.CreatedAtMs).UTC(),
	}
}
