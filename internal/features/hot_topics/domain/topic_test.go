
package domain

import (
	"testing"
	"time"
)

func TestTopic_NewTopicFromParts_GeneratesID(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	url := "https://example.com/x"
	topic, err := NewTopicFromParts(SourceWeibo, 1, " Hello  World ", &url, nil, nil, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic.ID != "weibo:https://example.com/x" {
		t.Fatalf("unexpected id: %q", topic.ID)
	}
	if topic.Title != " Hello  World " {
		t.Fatalf("title preserved, got %q", topic.Title)
	}
}

func TestTopic_NormalizeTitleAndID_NoURL(t *testing.T) {
	n := NormalizeTitle("  Hello	WORLD  ")
	if n != "hello world" {
		t.Fatalf("unexpected normalized: %q", n)
	}
	id := GenerateTopicID(SourceZhihu, "  Hello	WORLD  ", nil)
	if id != "zhihu:hello world" {
		t.Fatalf("unexpected id: %q", id)
	}
}

func TestTopic_ValidateErrors(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	_, err := NewTopicFromParts(SourceWeibo, 0, "t", nil, nil, nil, now)
	if err == nil {
		t.Fatalf("expected error")
	}
	_, err = NewTopicFromParts(SourceWeibo, 1, " ", nil, nil, nil, now)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	if !ContainsIgnoreCase("Hello World", "world") {
		t.Fatalf("expected match")
	}
	if ContainsIgnoreCase("Hello", "x") {
		t.Fatalf("expected no match")
	}
	if !ContainsIgnoreCase("Hello", " ") {
		t.Fatalf("empty query should match")
	}
}
