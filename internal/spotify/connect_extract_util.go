package spotify

import "strings"

func getMap(value any, path ...string) (map[string]any, bool) {
	current := value
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		next, ok := m[key]
		if !ok {
			return nil, false
		}
		current = next
	}
	m, ok := current.(map[string]any)
	return m, ok
}

func getString(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	if value, ok := m[key].(string); ok {
		return value
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	switch value := m[key].(type) {
	case int:
		return value
	case float64:
		return int(value)
	}
	return 0
}

func getNestedInt(m map[string]any, parent, key string) int {
	if nested, ok := getMap(m, parent); ok {
		return getInt(nested, key)
	}
	return 0
}

func getBool(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	if value, ok := m[key].(bool); ok {
		return value
	}
	return false
}

func getNestedBool(m map[string]any, parent, key string) bool {
	if nested, ok := getMap(m, parent); ok {
		return getBool(nested, key)
	}
	return false
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
