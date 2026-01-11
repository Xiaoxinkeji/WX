
package models

import (
	"database/sql"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type TopicDTO struct {
	ID          string
	Source      string
	Rank        int
	Title       string
	URL         sql.NullString
	HotValue    sql.NullFloat64
	Description sql.NullString
	FetchedAtMs int64
}

func (dto TopicDTO) ToDomain() (domain.Topic, error) {
	src, err := domain.ParseSource(dto.Source)
	if err != nil {
		return domain.Topic{}, err
	}

	var url *string
	if dto.URL.Valid {
		u := dto.URL.String
		url = &u
	}
	var hv *float64
	if dto.HotValue.Valid {
		v := dto.HotValue.Float64
		hv = &v
	}
	var desc *string
	if dto.Description.Valid {
		d := dto.Description.String
		desc = &d
	}

	t := domain.Topic{
		ID:          dto.ID,
		Source:      src,
		Rank:        dto.Rank,
		Title:       dto.Title,
		URL:         url,
		HotValue:    hv,
		Description: desc,
		FetchedAt:   time.UnixMilli(dto.FetchedAtMs).UTC(),
	}
	if err := t.Validate(); err != nil {
		return domain.Topic{}, err
	}
	return t, nil
}

func TopicDTOFromDomain(t domain.Topic) TopicDTO {
	dto := TopicDTO{
		ID:          t.ID,
		Source:      t.Source.Key(),
		Rank:        t.Rank,
		Title:       t.Title,
		FetchedAtMs: t.FetchedAt.UTC().UnixMilli(),
	}
	if t.URL != nil {
		dto.URL = sql.NullString{String: *t.URL, Valid: true}
	}
	if t.HotValue != nil {
		dto.HotValue = sql.NullFloat64{Float64: *t.HotValue, Valid: true}
	}
	if t.Description != nil {
		dto.Description = sql.NullString{String: *t.Description, Valid: true}
	}
	return dto
}
