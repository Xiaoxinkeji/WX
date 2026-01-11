package domain

import (
"errors"
"strings"
)

type Source string

const (
SourceWeibo Source = "weibo"
SourceZhihu Source = "zhihu"
SourceBaidu Source = "baidu"
SourceKr36  Source = "kr36"
)

func AllSources() []Source {
return []Source{SourceWeibo, SourceZhihu, SourceBaidu, SourceKr36}
}

func (s Source) Valid() bool {
switch s {
case SourceWeibo, SourceZhihu, SourceBaidu, SourceKr36:
return true
default:
return false
}
}

func (s Source) Key() string { return string(s) }

func (s Source) Label() string {
switch s {
case SourceWeibo:
return "??"
case SourceZhihu:
return "??"
case SourceBaidu:
return "??"
case SourceKr36:
return "36?"
default:
return string(s)
}
}

func ParseSource(key string) (Source, error) {
n := strings.TrimSpace(strings.ToLower(key))
s := Source(n)
if !s.Valid() {
return "", errors.New("hot_topics: invalid source")
}
return s, nil
}
