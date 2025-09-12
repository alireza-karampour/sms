package utils

import "strings"

type Subject string

func ErrContains(err error, txt string) bool {
	return strings.Contains(err.Error(), txt)
}

func MakeSubject(s ...string) string {
	return strings.Join(s, ".")
}

func (s Subject) Filter(subs ...string) bool {
	parts := strings.Split(string(s), ".")
	if len(parts) != len(subs) {
		return false
	}
	for v := range len(parts) {
		if subs[v] == "*" {
			continue
		}
		if parts[v] != subs[v] {
			return false
		}
	}
	return true
}
