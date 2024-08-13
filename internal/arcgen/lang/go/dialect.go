package arcgengo

import (
	"strconv"
	"strings"

	"github.com/kunitsucom/arcgen/internal/arcgen/lang/util"
	"github.com/kunitsucom/arcgen/internal/config"
)

//nolint:cyclop
func columnValuesPlaceholder(columns []string, initialNumber int) string {
	switch config.Dialect() {
	case "mysql", "sqlite3":
		// ?, ?, ?, ...
		return "?" + strings.Repeat(", ?", len(columns)-1)
	case "postgres", "cockroach":
		// $1, $2, $3, ...
		var s strings.Builder
		for i := range columns {
			if i > 0 {
				s.WriteString(", ")
			}
			s.WriteString("$" + strconv.Itoa(i+initialNumber))
		}
		return s.String()
	case "spanner":
		// @column_1, @column_2, @column_3, ...
		var s strings.Builder
		for i := range columns {
			if i > 0 {
				s.WriteString(", ")
			}
			s.WriteString("@" + columns[i])
		}
		return s.String()
	case "oracle":
		// :column_1, :column_2, :column_3, ...
		var s strings.Builder
		for i := range columns {
			if i > 0 {
				s.WriteString(", ")
			}
			s.WriteString(":" + columns[i])
		}
		return s.String()
	default:
		// ?, ?, ?, ...
		return "?" + strings.Repeat(", ?", len(columns)-1)
	}
}

//nolint:unparam,cyclop
func whereColumnsPlaceholder(columns []string, op string, initialNumber int) string {
	switch config.Dialect() {
	case "mysql", "sqlite3":
		// column1 = ? AND column2 = ? AND column3 = ...
		return util.JoinStringsWithQuote(columns, " = ? "+op+" ", quote) + " = ?"
	case "postgres", "cockroach":
		// column1 = $1 AND column2 = $2 AND column3 = ...
		var s strings.Builder
		for i, column := range columns {
			if i > 0 {
				s.WriteString(" " + op + " ")
			}
			s.WriteString(util.QuoteString(column, quote))
			s.WriteString(" = $")
			s.WriteString(strconv.Itoa(i + initialNumber))
		}
		return s.String()
	case "spanner":
		// column1 = @column_1 AND column2 = @column_2 AND column3 = ...
		var s strings.Builder
		for i, column := range columns {
			if i > 0 {
				s.WriteString(" " + op + " ")
			}
			s.WriteString(util.QuoteString(column, quote))
			s.WriteString(" = @")
			s.WriteString(column)
		}
		return s.String()
	case "oracle":
		// column1 = :column_1 AND column2 = :column_2 AND column3 = ...
		var s strings.Builder
		for i, column := range columns {
			if i > 0 {
				s.WriteString(" " + op + " ")
			}
			s.WriteString(util.QuoteString(column, quote))
			s.WriteString(" = :")
			s.WriteString(column)
		}
		return s.String()
	default:
		// column1 = ? AND column2 = ? AND column3 = ...
		return util.JoinStringsWithQuote(columns, " = ? "+op+" ", quote) + " = ?"
	}
}
