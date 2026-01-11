
package sources

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

func decodeJSON(body []byte) (any, error) {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func asMap(v any) map[string]any {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

func asList(v any) []any {
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	return list
}

func asString(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", false
	case string:
		return x, true
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64), true
	case bool:
		if x {
			return "true", true
		}
		return "false", true
	default:
		return "", false
	}
}

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case int:
		return x, true
	case int64:
		return int(x), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(x))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}

func asFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(x, ",", "")), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

var numberRE = regexp.MustCompile(`(\d+(?:\.\d+)?)`)

func tryParseFloatFromText(text string) (float64, bool) {
	m := numberRE.FindStringSubmatch(strings.ReplaceAll(text, ",", ""))
	if len(m) != 2 {
		return 0, false
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func maybeStringPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
