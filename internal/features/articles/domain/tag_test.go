package domain_test

import (
	"reflect"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

func TestNormalizeTagNames(t *testing.T) {
	got, err := domain.NormalizeTagNames([]string{"  Go  ", "go", "SQL", " sql "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"go", "sql"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	if _, err := domain.NormalizeTagNames([]string{""}); err == nil {
		t.Fatalf("expected error for empty tag")
	}
}
