package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound        = errors.New("articles: not found")
	ErrConflict        = errors.New("articles: conflict")
	ErrInvalidArgument = errors.New("articles: invalid argument")
)

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}

type CreateArticleParams struct {
	ID        string
	Title     string
	Content   string
	Status    ArticleStatus
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UpdateArticleParams struct {
	Title      *string
	Content    *string
	Status     *ArticleStatus
	Tags       *[]string
	UpdatedAt  time.Time
	IsAutoSave bool
}

type ListArticlesQuery struct {
	Status *ArticleStatus
	Tag    *string
	Limit  int
	Offset int
}

type SearchArticlesQuery struct {
	Query  string
	Status *ArticleStatus
	Tag    *string
	Limit  int
	Offset int
}

type ListVersionsQuery struct {
	ArticleID string
	Limit     int
	Offset    int
}

type ArticleCreator interface {
	CreateArticle(ctx context.Context, params CreateArticleParams) (Article, error)
}

type ArticleUpdater interface {
	UpdateArticle(ctx context.Context, articleID string, params UpdateArticleParams) (Article, error)
}

type ArticleDeleter interface {
	DeleteArticle(ctx context.Context, articleID string) error
}

type ArticleGetter interface {
	GetArticle(ctx context.Context, articleID string) (Article, error)
}

type ArticleLister interface {
	ListArticles(ctx context.Context, query ListArticlesQuery) ([]Article, error)
}

type ArticleSearcher interface {
	SearchArticles(ctx context.Context, query SearchArticlesQuery) ([]Article, error)
}

type TagLister interface {
	ListTags(ctx context.Context) ([]Tag, error)
}

type VersionLister interface {
	ListVersions(ctx context.Context, query ListVersionsQuery) ([]ArticleVersion, error)
	GetVersion(ctx context.Context, articleID string, version int) (ArticleVersion, error)
	RestoreVersion(ctx context.Context, articleID string, version int, restoredAt time.Time) (Article, error)
}

type Repository interface {
	ArticleCreator
	ArticleUpdater
	ArticleDeleter
	ArticleGetter
	ArticleLister
	ArticleSearcher
	TagLister
	VersionLister
}
