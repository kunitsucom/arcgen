package util

import (
	"strings"
)

func QuoteString(s string, quote string) string {
	return quote + s + quote
}

func JoinStringsWithQuote(ss []string, sep string, quote string) string {
	if len(ss) == 0 {
		return ""
	}

	if len(ss) == 1 {
		return QuoteString(ss[0], quote)
	}

	var builder strings.Builder
	for i, s := range ss {
		if i > 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(QuoteString(s, quote))
	}

	return builder.String()
}
