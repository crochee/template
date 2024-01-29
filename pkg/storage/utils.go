package storage

import "strings"

func EscapeQueryString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	replace := map[string]string{
		"%": `\%`,
		"_": `\_`,
		"'": `\'`,
		`"`: `\"`,
	}

	for old, n := range replace {
		s = strings.ReplaceAll(s, old, n)
	}
	return s
}
