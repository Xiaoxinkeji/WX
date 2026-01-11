package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// escapeFTS5Query escapes special characters in FTS5 queries to prevent syntax errors
func escapeFTS5Query(query string) string {
	// Remove or escape FTS5 special characters
	replacer := strings.NewReplacer(
		`"`, `""`,  // Escape double quotes
		`*`, ``,    // Remove wildcards
		`-`, ` `,   // Replace minus with space
		`(`, ``,    // Remove parentheses
		`)`, ``,
		`{`, ``,    // Remove braces
		`}`, ``,
		`[`, ``,    // Remove brackets
		`]`, ``,
		`^`, ``,    // Remove caret
		`:`, ``,    // Remove colon
	)
	escaped := replacer.Replace(query)

	// Trim and wrap in double quotes for phrase matching
	escaped = strings.TrimSpace(escaped)
	if escaped == "" {
		return `""`
	}

	// Use double quotes for exact phrase matching
	return `"` + escaped + `"`
}

type SQLiteSearchIndex struct {
	db *sql.DB
}

func NewSQLiteSearchIndex(db *sql.DB) (*SQLiteSearchIndex, error) {
	if db == nil {
		return nil, errors.New("search index: db is nil")
	}
	idx := &SQLiteSearchIndex{db: db}
	return idx, nil
}

func (s *SQLiteSearchIndex) EnsureSchema(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE VIRTUAL TABLE IF NOT EXISTS article_fts
USING fts5(title, content, tags, article_id UNINDEXED);
`)
	return err
}

func (s *SQLiteSearchIndex) UpsertTx(ctx context.Context, tx *sql.Tx, articleID, title, content string, tags []string) error {
	if tx == nil {
		return errors.New("search index: tx is nil")
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM article_fts WHERE article_id = ?`, articleID)
	if err != nil {
		return err
	}
	joinedTags := strings.Join(tags, " ")
	_, err = tx.ExecContext(ctx, `INSERT INTO article_fts(title, content, tags, article_id) VALUES(?, ?, ?, ?)`, title, content, joinedTags, articleID)
	return err
}

func (s *SQLiteSearchIndex) DeleteTx(ctx context.Context, tx *sql.Tx, articleID string) error {
	if tx == nil {
		return errors.New("search index: tx is nil")
	}
	_, err := tx.ExecContext(ctx, `DELETE FROM article_fts WHERE article_id = ?`, articleID)
	return err
}

func (s *SQLiteSearchIndex) SearchArticleIDs(ctx context.Context, query string, limit, offset int) ([]string, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Escape FTS5 special characters to prevent query syntax errors
	escapedQuery := escapeFTS5Query(q)

	rows, err := s.db.QueryContext(ctx, `
SELECT article_id
FROM article_fts
WHERE article_fts MATCH ?
ORDER BY bm25(article_fts)
LIMIT ? OFFSET ?
`, escapedQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
