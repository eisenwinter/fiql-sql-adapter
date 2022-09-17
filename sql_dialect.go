package fiqlsqladapter

import (
	"strconv"
	"strings"
)

type ParamStyle string

const PgParamStyle ParamStyle = "$"
const MssqlParamStyle ParamStyle = "@"
const StandardParamStyle ParamStyle = "?"

type DelimiterStyle string

const StandardSqlDelimiter DelimiterStyle = "\"\""
const MssqlDelimiter DelimiterStyle = "[]"
const MariaDelimiter DelimiterStyle = "``"
const NoDelimiter DelimiterStyle = ""

func delimitBuilder(style DelimiterStyle, col string, sb *strings.Builder) {
	switch style {
	case MssqlDelimiter:
		sb.WriteString("[")
	case MariaDelimiter:
		sb.WriteString("`")
	default:
		sb.WriteString(`"`)
	}
	sb.WriteString(col)
	switch style {
	case MssqlDelimiter:
		sb.WriteString("]")
	case MariaDelimiter:
		sb.WriteString("`")
	default:
		sb.WriteString(`"`)
	}
}

func parameterBuilder(style ParamStyle, len int, sb *strings.Builder) {
	switch style {
	case PgParamStyle:
		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(len))
	case MssqlParamStyle:
		sb.WriteString("@")
		sb.WriteString(strconv.Itoa(len))
	case StandardParamStyle:
		sb.WriteString("?")
	}
}

func WithDialectMSSQL() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = MssqlDelimiter
		a.paramStyle = MssqlParamStyle
	}
}

func WithDialectSQLite() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = StandardSqlDelimiter
		a.paramStyle = PgParamStyle
	}
}

func WithDialectPostgres() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = StandardSqlDelimiter
		a.paramStyle = PgParamStyle
	}
}

func WithDialectMariaDB() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = MariaDelimiter
		a.paramStyle = StandardParamStyle
	}
}
