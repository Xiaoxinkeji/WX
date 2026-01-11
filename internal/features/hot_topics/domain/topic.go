package domain

import (
"errors"
"fmt"
"regexp"
"strings"
"time"
)

type Topic struct {
ID          string
Source      Source
Rank        int
Title       string
URL         *string
HotValue    *float64
Description *string
FetchedAt   time.Time
}

func (t Topic) Validate() error {
if !t.Source.Valid() {
return errors.New("source is invalid")
}
if t.Rank <= 0 {
return errors.New("rank must be positive")
}
if strings.TrimSpace(t.Title) == "" {
return errors.New("title is required")
}
if t.FetchedAt.IsZero() {
return errors.New("fetchedAt is required")
}
if strings.TrimSpace(t.ID) == "" {
return errors.New("id is required")
}
return nil
}

func NewTopicFromParts(
source Source,
rank int,
title string,
url *string,
hotValue *float64,
description *string,
fetchedAt time.Time,
) (Topic, error) {
id := GenerateTopicID(source, title, url)
t := Topic{
ID:          id,
Source:      source,
Rank:        rank,
Title:       title,
URL:         url,
HotValue:    hotValue,
Description: description,
FetchedAt:   fetchedAt,
}
if err := t.Validate(); err != nil {
return Topic{}, err
}
return t, nil
}

var whitespaceRE = regexp.MustCompile(`\s+`)

func NormalizeTitle(title string) string {
n := strings.TrimSpace(strings.ToLower(title))
n = whitespaceRE.ReplaceAllString(n, " ")
return n
}

func GenerateTopicID(source Source, title string, url *string) string {
key := NormalizeTitle(title)
if url != nil {
if u := strings.TrimSpace(*url); u != "" {
key = u
}
}
return fmt.Sprintf("%s:%s", source.Key(), key)
}

func NormalizeQuery(query string) string {
return strings.TrimSpace(strings.ToLower(query))
}

func ContainsIgnoreCase(text, query string) bool {
q := NormalizeQuery(query)
if q == "" {
return true
}
return strings.Contains(strings.ToLower(text), q)
}
