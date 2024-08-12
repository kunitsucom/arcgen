package util

func PascalCaseToCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'A' && s[0] <= 'Z' {
		return string(s[0]+'a'-'A') + s[1:]
	}

	return s
}
