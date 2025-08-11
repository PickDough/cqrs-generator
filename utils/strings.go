package utils

import "strings"

func FirstLetterToUpper(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[0:1]) + s[1:]
}
