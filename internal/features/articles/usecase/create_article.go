package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type randomIDGenerator struct{}

func (randomIDGenerator) NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

type CreateArticleInput struct {
	Title   string
	Content string
	Status  domain.ArticleStatus
	Tags    []string
}

type CreateArticleUseCase struct {
	Repo  domain.ArticleCreator
	Clock domain.Clock
	IDs   domain.IDGenerator
}

func NewCreateArticleUseCase(repo domain.ArticleCreator) CreateArticleUseCase {
	return CreateArticleUseCase{Repo: repo, Clock: systemClock{}, IDs: randomIDGenerator{}}
}

func (uc CreateArticleUseCase) Execute(ctx context.Context, in CreateArticleInput) (domain.Article, error) {
	if uc.Repo == nil {
		return domain.Article{}, errors.New("create article: repo is nil")
	}
	if uc.Clock == nil {
		uc.Clock = systemClock{}
	}
	if uc.IDs == nil {
		uc.IDs = randomIDGenerator{}
	}

	status := in.Status
	if status == "" {
		status = domain.ArticleStatusDraft
	}
	if !status.Valid() {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("invalid status"))
	}

	normalizedTags, err := domain.NormalizeTagNames(in.Tags)
	if err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	if err := domain.ValidateArticleFields(status, in.Title, in.Content); err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	id, err := uc.IDs.NewID()
	if err != nil {
		return domain.Article{}, err
	}

	now := uc.Clock.Now()
	return uc.Repo.CreateArticle(ctx, domain.CreateArticleParams{
		ID:        id,
		Title:     in.Title,
		Content:   in.Content,
		Status:    status,
		Tags:      normalizedTags,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
