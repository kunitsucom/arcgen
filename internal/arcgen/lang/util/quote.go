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

	return quote + strings.Join(ss, quote+sep+quote) + quote
}
