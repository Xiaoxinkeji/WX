package data

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/data/models"
	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type SQLiteRepository struct {
	db    *sql.DB
	index *SQLiteSearchIndex
}

func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	if db == nil {
		return nil, errors.New("sqlite repository: db is nil")
	}
	idx, err := NewSQLiteSearchIndex(db)
	if err != nil {
		return nil, err
	}
	repo := &SQLiteRepository{db: db, index: idx}
	if err := repo.EnsureSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteRepository) EnsureSchema(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := r.db.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS articles (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at_ms INTEGER NOT NULL,
	updated_at_ms INTEGER NOT NULL,
	current_version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS article_versions (
	article_id TEXT NOT NULL,
	version INTEGER NOT NULL,
	title TEXT NOT NULL,
	content TEXT NOT NULL,
	status TEXT NOT NULL,
	tags_csv TEXT NOT NULL,
	created_at_ms INTEGER NOT NULL,
	is_autosave INTEGER NOT NULL,
	PRIMARY KEY(article_id, version),
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tags (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	created_at_ms INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS article_tags (
	article_id TEXT NOT NULL,
	tag_id TEXT NOT NULL,
	PRIMARY KEY(article_id, tag_id),
	FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE,
	FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_articles_updated_at_ms ON articles(updated_at_ms DESC);
CREATE INDEX IF NOT EXISTS idx_article_tags_article_id ON article_tags(article_id);
CREATE INDEX IF NOT EXISTS idx_article_tags_tag_id ON article_tags(tag_id);
`); err != nil {
		return err
	}
	return r.index.EnsureSchema(ctx)
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *SQLiteRepository) CreateArticle(ctx context.Context, params domain.CreateArticleParams) (domain.Article, error) {
	if err := domain.ValidateArticleFields(params.Status, params.Title, params.Content); err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}
	if params.ID == "" {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}

	normalizedTags, err := domain.NormalizeTagNames(params.Tags)
	if err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	createdAtMs := params.CreatedAt.UTC().UnixMilli()
	updatedAtMs := params.UpdatedAt.UTC().UnixMilli()
	if createdAtMs == 0 {
		createdAtMs = time.Now().UTC().UnixMilli()
	}
	if updatedAtMs == 0 {
		updatedAtMs = createdAtMs
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Article{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return domain.Article{}, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO articles(id, title, content, status, created_at_ms, updated_at_ms, current_version)
VALUES(?, ?, ?, ?, ?, ?, 1)
`, params.ID, params.Title, params.Content, string(params.Status), createdAtMs, updatedAtMs); err != nil {
		if isUniqueConstraintErr(err) {
			return domain.Article{}, errors.Join(domain.ErrConflict, err)
		}
		return domain.Article{}, err
	}

	tags, err := r.replaceTagsTx(ctx, tx, params.ID, normalizedTags, createdAtMs)
	if err != nil {
		return domain.Article{}, err
	}

	tagsCSV := strings.Join(normalizedTags, ",")
	if _, err := tx.ExecContext(ctx, `
INSERT INTO article_versions(article_id, version, title, content, status, tags_csv, created_at_ms, is_autosave)
VALUES(?, 1, ?, ?, ?, ?, ?, 0)
`, params.ID, params.Title, params.Content, string(params.Status), tagsCSV, updatedAtMs); err != nil {
		return domain.Article{}, err
	}

	if err := r.index.UpsertTx(ctx, tx, params.ID, params.Title, params.Content, normalizedTags); err != nil {
		return domain.Article{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Article{}, err
	}

	return (models.ArticleDTO{
		ID:             params.ID,
		Title:          params.Title,
		Content:        params.Content,
		Status:         string(params.Status),
		CreatedAtMs:    createdAtMs,
		UpdatedAtMs:    updatedAtMs,
		CurrentVersion: 1,
	}).ToDomain(tags)
}

func (r *SQLiteRepository) GetArticle(ctx context.Context, articleID string) (domain.Article, error) {
	if articleID == "" {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}

	var dto models.ArticleDTO
	if err := r.db.QueryRowContext(ctx, `
SELECT id, title, content, status, created_at_ms, updated_at_ms, current_version
FROM articles
WHERE id = ?
`, articleID).Scan(&dto.ID, &dto.Title, &dto.Content, &dto.Status, &dto.CreatedAtMs, &dto.UpdatedAtMs, &dto.CurrentVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrNotFound
		}
		return domain.Article{}, err
	}

	tags, err := r.fetchTags(ctx, r.db, articleID)
	if err != nil {
		return domain.Article{}, err
	}
	return dto.ToDomain(tags)
}

func (r *SQLiteRepository) UpdateArticle(ctx context.Context, articleID string, params domain.UpdateArticleParams) (domain.Article, error) {
	if articleID == "" {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Article{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return domain.Article{}, err
	}

	var existing models.ArticleDTO
	if err := tx.QueryRowContext(ctx, `
SELECT id, title, content, status, created_at_ms, updated_at_ms, current_version
FROM articles
WHERE id = ?
`, articleID).Scan(&existing.ID, &existing.Title, &existing.Content, &existing.Status, &existing.CreatedAtMs, &existing.UpdatedAtMs, &existing.CurrentVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrNotFound
		}
		return domain.Article{}, err
	}

	newTitle := existing.Title
	if params.Title != nil {
		newTitle = *params.Title
	}
	newContent := existing.Content
	if params.Content != nil {
		newContent = *params.Content
	}
	newStatus := domain.ArticleStatus(existing.Status)
	if params.Status != nil {
		newStatus = *params.Status
	}

	var tagNames []string
	if params.Tags != nil {
		normalized, err := domain.NormalizeTagNames(*params.Tags)
		if err != nil {
			return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
		}
		tagNames = normalized
	} else {
		existingTags, err := r.fetchTagNames(ctx, tx, articleID)
		if err != nil {
			return domain.Article{}, err
		}
		tagNames = existingTags
	}

	if err := domain.ValidateArticleFields(newStatus, newTitle, newContent); err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	updatedAtMs := params.UpdatedAt.UTC().UnixMilli()
	if updatedAtMs == 0 {
		updatedAtMs = time.Now().UTC().UnixMilli()
	}

	newVersion := existing.CurrentVersion + 1
	if _, err := tx.ExecContext(ctx, `
UPDATE articles
SET title = ?, content = ?, status = ?, updated_at_ms = ?, current_version = ?
WHERE id = ?
`, newTitle, newContent, string(newStatus), updatedAtMs, newVersion, articleID); err != nil {
		return domain.Article{}, err
	}

	var tags []domain.Tag
	if params.Tags != nil {
		tags, err = r.replaceTagsTx(ctx, tx, articleID, tagNames, updatedAtMs)
		if err != nil {
			return domain.Article{}, err
		}
	} else {
		tags, err = r.fetchTags(ctx, tx, articleID)
		if err != nil {
			return domain.Article{}, err
		}
	}

	tagsCSV := strings.Join(tagNames, ",")
	isAutoSave := 0
	if params.IsAutoSave {
		isAutoSave = 1
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO article_versions(article_id, version, title, content, status, tags_csv, created_at_ms, is_autosave)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, articleID, newVersion, newTitle, newContent, string(newStatus), tagsCSV, updatedAtMs, isAutoSave); err != nil {
		return domain.Article{}, err
	}

	if err := r.index.UpsertTx(ctx, tx, articleID, newTitle, newContent, tagNames); err != nil {
		return domain.Article{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Article{}, err
	}

	return (models.ArticleDTO{
		ID:             articleID,
		Title:          newTitle,
		Content:        newContent,
		Status:         string(newStatus),
		CreatedAtMs:    existing.CreatedAtMs,
		UpdatedAtMs:    updatedAtMs,
		CurrentVersion: newVersion,
	}).ToDomain(tags)
}

func (r *SQLiteRepository) DeleteArticle(ctx context.Context, articleID string) error {
	if articleID == "" {
		return errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}
	_tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer _tx.Rollback()

	if _, err := _tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return err
	}

	if err := r.index.DeleteTx(ctx, _tx, articleID); err != nil {
		return err
	}

	res, err := _tx.ExecContext(ctx, `DELETE FROM articles WHERE id = ?`, articleID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}

	return _tx.Commit()
}

func (r *SQLiteRepository) ListArticles(ctx context.Context, query domain.ListArticlesQuery) ([]domain.Article, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	ids, err := r.listArticleIDs(ctx, query.Status, query.Tag, limit, offset)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return r.getArticlesByIDs(ctx, ids, query.Status, query.Tag)
}

func (r *SQLiteRepository) SearchArticles(ctx context.Context, query domain.SearchArticlesQuery) ([]domain.Article, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	ids, err := r.index.SearchArticleIDs(ctx, query.Query, limit, offset)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}
	return r.getArticlesByIDs(ctx, ids, query.Status, query.Tag)
}

func (r *SQLiteRepository) ListTags(ctx context.Context) ([]domain.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at_ms FROM tags ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Tag, 0)
	for rows.Next() {
		var dto models.TagDTO
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.CreatedAtMs); err != nil {
			return nil, err
		}
		out = append(out, dto.ToDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) ListVersions(ctx context.Context, query domain.ListVersionsQuery) ([]domain.ArticleVersion, error) {
	if query.ArticleID == "" {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT article_id, version, title, content, status, tags_csv, created_at_ms, is_autosave
FROM article_versions
WHERE article_id = ?
ORDER BY version DESC
LIMIT ? OFFSET ?
`, query.ArticleID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ArticleVersion, 0)
	for rows.Next() {
		var dto models.ArticleVersionDTO
		if err := rows.Scan(&dto.ArticleID, &dto.Version, &dto.Title, &dto.Content, &dto.Status, &dto.TagsCSV, &dto.CreatedAtMs, &dto.IsAutoSave); err != nil {
			return nil, err
		}
		v, err := dto.ToDomain()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) GetVersion(ctx context.Context, articleID string, version int) (domain.ArticleVersion, error) {
	if articleID == "" {
		return domain.ArticleVersion{}, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}
	if version <= 0 {
		return domain.ArticleVersion{}, errors.Join(domain.ErrInvalidArgument, errors.New("version must be positive"))
	}

	var dto models.ArticleVersionDTO
	if err := r.db.QueryRowContext(ctx, `
SELECT article_id, version, title, content, status, tags_csv, created_at_ms, is_autosave
FROM article_versions
WHERE article_id = ? AND version = ?
`, articleID, version).Scan(&dto.ArticleID, &dto.Version, &dto.Title, &dto.Content, &dto.Status, &dto.TagsCSV, &dto.CreatedAtMs, &dto.IsAutoSave); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ArticleVersion{}, domain.ErrNotFound
		}
		return domain.ArticleVersion{}, err
	}
	v, err := dto.ToDomain()
	if err != nil {
		return domain.ArticleVersion{}, err
	}
	return v, nil
}

func (r *SQLiteRepository) RestoreVersion(ctx context.Context, articleID string, version int, restoredAt time.Time) (domain.Article, error) {
	if articleID == "" {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}
	if version <= 0 {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("version must be positive"))
	}
	if restoredAt.IsZero() {
		restoredAt = time.Now().UTC()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Article{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return domain.Article{}, err
	}

	var existing models.ArticleDTO
	if err := tx.QueryRowContext(ctx, `
SELECT id, title, content, status, created_at_ms, updated_at_ms, current_version
FROM articles
WHERE id = ?
`, articleID).Scan(&existing.ID, &existing.Title, &existing.Content, &existing.Status, &existing.CreatedAtMs, &existing.UpdatedAtMs, &existing.CurrentVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrNotFound
		}
		return domain.Article{}, err
	}

	var vdto models.ArticleVersionDTO
	if err := tx.QueryRowContext(ctx, `
SELECT article_id, version, title, content, status, tags_csv, created_at_ms, is_autosave
FROM article_versions
WHERE article_id = ? AND version = ?
`, articleID, version).Scan(&vdto.ArticleID, &vdto.Version, &vdto.Title, &vdto.Content, &vdto.Status, &vdto.TagsCSV, &vdto.CreatedAtMs, &vdto.IsAutoSave); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Article{}, domain.ErrNotFound
		}
		return domain.Article{}, err
	}
	ver, err := vdto.ToDomain()
	if err != nil {
		return domain.Article{}, err
	}

	newStatus := ver.Status
	if err := domain.ValidateArticleFields(newStatus, ver.Title, ver.Content); err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	newVersion := existing.CurrentVersion + 1
	updatedAtMs := restoredAt.UTC().UnixMilli()
	if _, err := tx.ExecContext(ctx, `
UPDATE articles
SET title = ?, content = ?, status = ?, updated_at_ms = ?, current_version = ?
WHERE id = ?
`, ver.Title, ver.Content, string(newStatus), updatedAtMs, newVersion, articleID); err != nil {
		return domain.Article{}, err
	}

	normalizedTags, err := domain.NormalizeTagNames(ver.Tags)
	if err != nil {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
	}
	tags, err := r.replaceTagsTx(ctx, tx, articleID, normalizedTags, updatedAtMs)
	if err != nil {
		return domain.Article{}, err
	}

	tagsCSV := strings.Join(normalizedTags, ",")
	if _, err := tx.ExecContext(ctx, `
INSERT INTO article_versions(article_id, version, title, content, status, tags_csv, created_at_ms, is_autosave)
VALUES(?, ?, ?, ?, ?, ?, ?, 0)
`, articleID, newVersion, ver.Title, ver.Content, string(newStatus), tagsCSV, updatedAtMs); err != nil {
		return domain.Article{}, err
	}

	if err := r.index.UpsertTx(ctx, tx, articleID, ver.Title, ver.Content, normalizedTags); err != nil {
		return domain.Article{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Article{}, err
	}

	return (models.ArticleDTO{
		ID:             articleID,
		Title:          ver.Title,
		Content:        ver.Content,
		Status:         string(newStatus),
		CreatedAtMs:    existing.CreatedAtMs,
		UpdatedAtMs:    updatedAtMs,
		CurrentVersion: newVersion,
	}).ToDomain(tags)
}

func (r *SQLiteRepository) listArticleIDs(ctx context.Context, status *domain.ArticleStatus, tag *string, limit, offset int) ([]string, error) {
	var b strings.Builder
	args := make([]any, 0, 4)
	b.WriteString("SELECT a.id FROM articles a")
	if tag != nil {
		b.WriteString(" JOIN article_tags at ON at.article_id = a.id JOIN tags t ON t.id = at.tag_id")
	}
	b.WriteString(" WHERE 1=1")
	if status != nil {
		b.WriteString(" AND a.status = ?")
		args = append(args, string(*status))
	}
	if tag != nil {
		b.WriteString(" AND t.name = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(*tag)))
	}
	b.WriteString(" ORDER BY a.updated_at_ms DESC LIMIT ? OFFSET ?")
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) getArticlesByIDs(ctx context.Context, ids []string, status *domain.ArticleStatus, tag *string) ([]domain.Article, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	args := make([]any, 0, len(ids)*2+4)
	var b strings.Builder
	b.WriteString("SELECT a.id, a.title, a.content, a.status, a.created_at_ms, a.updated_at_ms, a.current_version FROM articles a")
	if tag != nil {
		b.WriteString(" JOIN article_tags at ON at.article_id = a.id JOIN tags t ON t.id = at.tag_id")
	}
	b.WriteString(" WHERE a.id IN (")
	b.WriteString(placeholders(len(ids)))
	b.WriteString(")")
	for _, id := range ids {
		args = append(args, id)
	}
	if status != nil {
		b.WriteString(" AND a.status = ?")
		args = append(args, string(*status))
	}
	if tag != nil {
		b.WriteString(" AND t.name = ?")
		args = append(args, strings.ToLower(strings.TrimSpace(*tag)))
	}
	b.WriteString(" ORDER BY ")
	b.WriteString(caseOrder("a.id", ids))
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dtos := make([]models.ArticleDTO, 0, len(ids))
	returnedIDs := make([]string, 0, len(ids))
	for rows.Next() {
		var dto models.ArticleDTO
		if err := rows.Scan(&dto.ID, &dto.Title, &dto.Content, &dto.Status, &dto.CreatedAtMs, &dto.UpdatedAtMs, &dto.CurrentVersion); err != nil {
			return nil, err
		}
		dtos = append(dtos, dto)
		returnedIDs = append(returnedIDs, dto.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(dtos) == 0 {
		return nil, nil
	}

	tagsByArticle, err := r.fetchTagsByArticles(ctx, r.db, returnedIDs)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Article, 0, len(dtos))
	for _, dto := range dtos {
		article, err := dto.ToDomain(tagsByArticle[dto.ID])
		if err != nil {
			return nil, err
		}
		out = append(out, article)
	}
	return out, nil
}

func (r *SQLiteRepository) fetchTags(ctx context.Context, q queryer, articleID string) ([]domain.Tag, error) {
	rows, err := q.QueryContext(ctx, `
SELECT t.id, t.name, t.created_at_ms
FROM tags t
JOIN article_tags at ON at.tag_id = t.id
WHERE at.article_id = ?
ORDER BY t.name ASC
`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Tag, 0)
	for rows.Next() {
		var dto models.TagDTO
		if err := rows.Scan(&dto.ID, &dto.Name, &dto.CreatedAtMs); err != nil {
			return nil, err
		}
		out = append(out, dto.ToDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) fetchTagNames(ctx context.Context, q queryer, articleID string) ([]string, error) {
	rows, err := q.QueryContext(ctx, `
SELECT t.name
FROM tags t
JOIN article_tags at ON at.tag_id = t.id
WHERE at.article_id = ?
ORDER BY t.name ASC
`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) fetchTagsByArticles(ctx context.Context, q queryer, articleIDs []string) (map[string][]domain.Tag, error) {
	if len(articleIDs) == 0 {
		return map[string][]domain.Tag{}, nil
	}

	args := make([]any, 0, len(articleIDs))
	for _, id := range articleIDs {
		args = append(args, id)
	}

	rows, err := q.QueryContext(ctx, fmt.Sprintf(`
SELECT at.article_id, t.id, t.name, t.created_at_ms
FROM article_tags at
JOIN tags t ON t.id = at.tag_id
WHERE at.article_id IN (%s)
ORDER BY t.name ASC
`, placeholders(len(articleIDs))), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]domain.Tag, len(articleIDs))
	for rows.Next() {
		var articleID string
		var dto models.TagDTO
		if err := rows.Scan(&articleID, &dto.ID, &dto.Name, &dto.CreatedAtMs); err != nil {
			return nil, err
		}
		out[articleID] = append(out[articleID], dto.ToDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) replaceTagsTx(ctx context.Context, tx *sql.Tx, articleID string, tagNames []string, nowMs int64) ([]domain.Tag, error) {
	if _, err := tx.ExecContext(ctx, `DELETE FROM article_tags WHERE article_id = ?`, articleID); err != nil {
		return nil, err
	}
	if len(tagNames) == 0 {
		return nil, nil
	}

	out := make([]domain.Tag, 0, len(tagNames))
	for _, name := range tagNames {
		tag, err := upsertTagTx(ctx, tx, name, nowMs)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO article_tags(article_id, tag_id) VALUES(?, ?)`, articleID, tag.ID); err != nil {
			return nil, err
		}
		out = append(out, tag)
	}
	return out, nil
}

func upsertTagTx(ctx context.Context, tx *sql.Tx, normalizedName string, nowMs int64) (domain.Tag, error) {
	var dto models.TagDTO
	if err := tx.QueryRowContext(ctx, `SELECT id, name, created_at_ms FROM tags WHERE name = ?`, normalizedName).Scan(&dto.ID, &dto.Name, &dto.CreatedAtMs); err == nil {
		return dto.ToDomain(), nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return domain.Tag{}, err
	}

	id, err := newRandomID()
	if err != nil {
		return domain.Tag{}, err
	}
	if nowMs == 0 {
		nowMs = time.Now().UTC().UnixMilli()
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO tags(id, name, created_at_ms) VALUES(?, ?, ?)`, id, normalizedName, nowMs); err != nil {
		if isUniqueConstraintErr(err) {
			// Another transaction inserted it; re-select.
			if err2 := tx.QueryRowContext(ctx, `SELECT id, name, created_at_ms FROM tags WHERE name = ?`, normalizedName).Scan(&dto.ID, &dto.Name, &dto.CreatedAtMs); err2 != nil {
				return domain.Tag{}, err2
			}
			return dto.ToDomain(), nil
		}
		return domain.Tag{}, err
	}
	return (models.TagDTO{ID: id, Name: normalizedName, CreatedAtMs: nowMs}).ToDomain(), nil
}

func newRandomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := strings.Repeat("?,", n)
	return b[:len(b)-1]
}

func caseOrder(column string, ids []string) string {
	var b strings.Builder
	b.WriteString("CASE ")
	b.WriteString(column)
	for i := range ids {
		b.WriteString(" WHEN ? THEN ")
		b.WriteString(fmt.Sprintf("%d", i))
	}
	b.WriteString(" ELSE ")
	b.WriteString(fmt.Sprintf("%d", len(ids)))
	b.WriteString(" END")
	return b.String()
}

func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint") || strings.Contains(msg, "constraint failed")
}
