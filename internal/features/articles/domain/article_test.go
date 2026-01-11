package domain_test

import (
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

func TestValidateArticleFields(t *testing.T) {
	if err := domain.ValidateArticleFields(domain.ArticleStatusDraft, "", ""); err != nil {
		t.Fatalf("draft should allow empty title/content: %v", err)
	}
	if err := domain.ValidateArticleFields(domain.ArticleStatusPublished, "", "x"); err == nil {
		t.Fatalf("published should require title")
	}
	if err := domain.ValidateArticleFields(domain.ArticleStatusPublished, "t", ""); err == nil {
		t.Fatalf("published should require content")
	}
	if err := domain.ValidateArticleFields(domain.ArticleStatusPublished, "t", "c"); err != nil {
		t.Fatalf("published should be valid: %v", err)
	}
	if err := domain.ValidateArticleFields(domain.ArticleStatus("bad"), "t", "c"); err == nil {
		t.Fatalf("invalid status should fail")
	}
}
