package arcgengo

import (
	"strconv"
	"strings"

	"github.com/kunitsucom/arcgen/internal/config"
)

func columnValuesPlaceholder(columns []string) string {
	switch config.Dialect() {
	case "mysql", "sqlite3":
		// ?, ?, ?, ...
		return "?" + strings.Repeat(", ?", len(columns)-1)
	case "postgres", "cockroach":
		// $1, $2, $3, ...
		var s strings.Builder
		s.WriteString("$1")
		for i := 2; i <= len(columns); i++ {
			s.WriteString(", $")
			s.WriteString(strconv.Itoa(i))
		}
		return s.String()
	case "spanner":
		// @column_1, @column_2, @column_3, ...
		var s strings.Builder
		s.WriteString("@" + columns[0])
		for i := 2; i <= len(columns); i++ {
			s.WriteString(", @")
			s.WriteString(columns[i-1])
		}
		return s.String()
	case "oracle":
		// :column_1, :column_2, :column_3, ...
		var s strings.Builder
		s.WriteString(":" + columns[0])
		for i := 2; i <= len(columns); i++ {
			s.WriteString(", :")
			s.WriteString(columns[i-1])
		}
		return s.String()
	default:
		// ?, ?, ?, ...
		return "?" + strings.Repeat(", ?", len(columns)-1)
	}
}

//nolint:unparam,cyclop
func whereColumnsPlaceholder(columns []string, op string) string {
	switch config.Dialect() {
	case "mysql", "sqlite3":
		// column1 = ? AND column2 = ? AND column3 = ...
		return strings.Join(columns, " = ? "+op+" ") + " = ?"
	case "postgres", "cockroach":
		// column1 = $1 AND column2 = $2 AND column3 = ...
		var s strings.Builder
		for i, column := range columns {
			if i > 0 {
				s.WriteString(" " + op + " ")
			}
			s.WriteString(column)
			s.WriteString(" = $")
			s.WriteString(strconv.Itoa(i + 1))
		}
		return s.String()
	case "spanner":
		// column1 = @column_1 AND column2 = @column_2 AND column3 = ...
		var s strings.Builder
		for i, column := range columns {
			if i > 0 {
				s.WriteString(" " + op + " ")
			}
			s.WriteString(column)
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
			s.WriteString(column)
			s.WriteString(" = :")
			s.WriteString(column)
		}
		return s.String()
	default:
		// column1 = ? AND column2 = ? AND column3 = ...
		return strings.Join(columns, " = ? "+op+" ") + " = ?"
	}
}
