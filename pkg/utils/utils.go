package utils

import "strings"

func ErrContains(err error, txt string) bool {
	return strings.Contains(err.Error(), txt)
}
