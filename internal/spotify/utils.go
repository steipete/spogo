package spotify

import "strings"

func joinComma(values []string) string {
	return strings.Join(values, ",")
}
