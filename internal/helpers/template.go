package helpers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"mhp-rooms/internal/config"
	"mhp-rooms/internal/utils"
)

func toLower(s string) string {
	return strings.ToLower(s)
}

func toJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}

// makeMap テンプレート内で動的にマップを構築
func makeMap(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("makeMap called with odd number of arguments")
	}
	m := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("makeMap keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m, nil
}

func gameVersionIconHTML(code string) template.HTML {
	return template.HTML(utils.GetGameVersionIcon(code))
}

// safeString nil安全な文字列変換
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func hasStringValue(s *string) bool {
	return s != nil && *s != ""
}

func jsEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func jsEscapePtr(s *string) string {
	if s == nil {
		return ""
	}
	return jsEscape(*s)
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}

func mul(a, b int) int {
	return a * b
}

// min intとint64の両方に対応した最小値取得
func min(a, b interface{}) int64 {
	var aVal, bVal int64

	switch v := a.(type) {
	case int:
		aVal = int64(v)
	case int64:
		aVal = v
	default:
		aVal = 0
	}

	switch v := b.(type) {
	case int:
		bVal = int64(v)
	case int64:
		bVal = v
	default:
		bVal = 0
	}

	if aVal < bVal {
		return aVal
	}
	return bVal
}

func sequence(start, end int) []int {
	result := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		result = append(result, i)
	}
	return result
}

func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"lower":            toLower,
		"json":             toJSON,
		"map":              makeMap,
		"gameVersionColor": utils.GetGameVersionColor,
		"gameVersionIcon":  gameVersionIconHTML,
		"gameVersionAbbr":  utils.GetGameVersionAbbreviation,
		"stringPtr":        safeString,
		"safeString":       safeString,
		"hasStringValue":   hasStringValue,
		"jsEscape":         jsEscape,
		"jsEscapePtr":      jsEscapePtr,
		"add":              add,
		"sub":              sub,
		"mul":              mul,
		"min":              min,
		"sequence":         sequence,
		"getEnv":           config.GetEnv,
	}
}
