package fiqlsqladapter

import (
	"strconv"
	"strings"
)

// paramStyle defines how parameters look like in the selected sql dialect
type paramStyle string

// dollarParamStyle is used by postgres (and sqlite)
const dollarParamStyle paramStyle = "$"

// atParamStyle is used by mssql (and sqlite yeah really ...)
const atParamStyle paramStyle = "@"

// standardParamStyle is the standard ? (and yeah sqlite supports that too...)
const standardParamStyle paramStyle = "?"

// delimiterStyle defines the sql column delimiter
type delimiterStyle string

// standardSqlDelimiter as defined in the SQL92 standard
const standardSqlDelimiter delimiterStyle = "\"\""

// angleBracketDelimiter is used by mssql
const angleBracketDelimiter delimiterStyle = "[]"

// backtickDelimiter is used by mariadb and mysql
const backtickDelimiter delimiterStyle = "``"

// noDelimiter simply means none
const noDelimiter delimiterStyle = ""

// concatSupport defines if the database has a concat function or another mean of concatinating strings
type concatSupport string

// concatFunctionSupported indicates that the database system supports the CONCAT(1..n) function
const concatFunctionSupported concatSupport = "concat"

// concatByPipesSupported inidicates the database uses the double pipe operator for string concatination
const concatByPipesSupported concatSupport = "||"

func delimitBuilder(style delimiterStyle, col string, sb *strings.Builder) {
	switch style {
	case angleBracketDelimiter:
		sb.WriteString("[")
	case backtickDelimiter:
		sb.WriteString("`")
	default:
		sb.WriteString(`"`)
	}
	sb.WriteString(col)
	switch style {
	case angleBracketDelimiter:
		sb.WriteString("]")
	case backtickDelimiter:
		sb.WriteString("`")
	default:
		sb.WriteString(`"`)
	}
}

func parameterBuilder(style paramStyle, len int, sb *strings.Builder) {
	switch style {
	case dollarParamStyle:
		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(len))
	case atParamStyle:
		sb.WriteString("@")
		sb.WriteString(strconv.Itoa(len))
	case standardParamStyle:
		sb.WriteString("?")
	}
}

// Dialect constants
const (
	DialectMariaDB  = "maria"
	DialectMySQL    = "mysql"
	DialectMSSL     = "mssql"
	DialectSQLite   = "sqlite3"
	DialectPostgres = "postgres"
)

// WithDialect configures which dialect to be used
func WithDialect(dialect string) func(*Adapter) {
	switch dialect {
	case DialectMariaDB:
	case DialectMySQL:
		return WithDialectMariaDB()
	case DialectPostgres:
		return WithDialectPostgres()
	case DialectMSSL:
		return WithDialectMSSQL()
	case DialectSQLite:
		return WithDialectSQLite()
	}
	return WithDialectSQL92()
}

// WithDialectMSSQL configures MSSQL delimiters and params
func WithDialectMSSQL() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = angleBracketDelimiter
		a.paramStyle = atParamStyle
		a.concat = concatFunctionSupported
	}
}

// WithDialectSQLite configures SQLite delimiters and params
func WithDialectSQLite() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = standardSqlDelimiter
		a.paramStyle = standardParamStyle
		a.concat = concatByPipesSupported
	}
}

// WithDialectPostgres configures Postgres delimiters and params
func WithDialectPostgres() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = standardSqlDelimiter
		a.paramStyle = dollarParamStyle
		a.concat = concatFunctionSupported
	}
}

// WithDialectMariaDB configures MariaDB / MySql delimiters and params
func WithDialectMariaDB() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = backtickDelimiter
		a.paramStyle = standardParamStyle
		a.concat = concatFunctionSupported
	}
}

// WithDialectSQL92 means column delimiter is " and parameters are ?
func WithDialectSQL92() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = standardSqlDelimiter
		a.paramStyle = standardParamStyle
		a.concat = concatFunctionSupported
	}
}

// WithDialectSQL92NoDelimiter means no column delimiter is used and parameters are ?
func WithDialectSQL92NoDelimiter() func(*Adapter) {
	return func(a *Adapter) {
		a.delim = noDelimiter
		a.paramStyle = standardParamStyle
		a.concat = concatFunctionSupported
	}
}
