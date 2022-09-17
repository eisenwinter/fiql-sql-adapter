package fiqlsqladapter

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	fq "github.com/eisenwinter/fiql-parser"
)

var stringType = reflect.TypeOf("")
var timeType = reflect.TypeOf(time.Time{})

var float64Type = reflect.TypeOf(float64(0))
var intType = reflect.TypeOf(int(0))

type FieldMapping map[string]Field
type ParamStyle string

const PgParamStyle ParamStyle = "$"
const MssqlParamStyle ParamStyle = "@"
const StandardParamStyle ParamStyle = "?"

// consider case sensitivity
// https://en.wikibooks.org/wiki/SQL_Dialects_Reference/Print_version

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

func concatErrrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	}
	err := errs[0]
	for _, e := range errs[1:] {
		err = fmt.Errorf("%w", e)
	}
	return err
}

type Adapter struct {
	fields     FieldMapping
	parser     *fq.Parser
	delim      DelimiterStyle
	paramStyle ParamStyle
}

type whereBuilder struct {
	sb           strings.Builder
	params       []interface{}
	errors       []error
	lastSelector *Field
	fields       FieldMapping
	delim        DelimiterStyle
	paramStyle   ParamStyle
}

func (t *whereBuilder) VisitExpressionEntered() { t.sb.WriteString("(") }
func (t *whereBuilder) VisitExpressionLeft()    { t.sb.WriteString(")") }
func (t *whereBuilder) VisitOperator(operatorCtx fq.OperatorContext) {
	switch operatorCtx.Operator() {
	case fq.OperatorAND:
		t.sb.WriteString(" AND ")
	case fq.OperatorOR:
		t.sb.WriteString(" OR ")
	}
}

func (t *whereBuilder) VisitSelector(selectorCtx fq.SelectorContext) {
	selector := selectorCtx.Selector()
	if fi, ok := t.fields[strings.ToLower(selector)]; ok {
		delimitBuilder(t.delim, fi.Db, &t.sb)
		if selectorCtx.IsUnary() {
			t.lastSelector = nil
			t.sb.WriteString(" IS NOT NULL")
		} else {
			t.lastSelector = &fi
		}

	} else {
		t.errors = append(t.errors, fmt.Errorf("invalid selector: %s", selector))
		t.lastSelector = nil
	}

}
func (t *whereBuilder) VisitComparison(comparisonCtx fq.ComparisonContext) {
	if t.lastSelector == nil {
		return
	}
	switch comparisonCtx.Comparison() {
	case fq.ComparisonEq:
		if t.lastSelector.Type == stringType {
			t.sb.WriteString(" LIKE ")
		} else {
			t.sb.WriteString(" = ")
		}

	case fq.ComparisonNeq:
		t.sb.WriteString(" <> ")
	case fq.ComparisonGt:
		t.sb.WriteString(" > ")
	case fq.ComparisonGte:
		t.sb.WriteString(" >= ")
	case fq.ComparisonLt:
		t.sb.WriteString(" < ")
	case fq.ComparisonLte:
		t.sb.WriteString(" <= ")
	}
}

func (t *whereBuilder) isCompatibleType(from, to reflect.Type) bool {
	for to.Kind() == reflect.Ptr || to.Kind() == reflect.Interface {
		to = to.Elem()
	}
	return from == to
}

func (t *whereBuilder) negotiateArgumentType(args *fq.ArgumentContext) (bool, error) {
	exp := t.lastSelector.Type
	if args.ValueRecommendation() == fq.ValueRecommendationString && exp == stringType {
		//its safe to assume that string is a string
		t.params = append(t.params, args.AsString())
		return true, nil
	}

	switch args.ValueRecommendation() {
	case fq.ValueRecommendationDateTime:
		if t.isCompatibleType(timeType, exp) {
			time, err := args.AsTime()
			if err != nil {
				return false, err
			}
			t.params = append(t.params, time)
			return false, nil
		}
		break
	case fq.ValueRecommendationDuration:
		if t.isCompatibleType(timeType, exp) {
			time, err := args.AsDuration()
			if err != nil {
				return false, err
			}
			t.params = append(t.params, time)
			return false, nil
		}
		break
	case fq.ValueRecommendationNumber:
		// is it int
		if t.isCompatibleType(intType, exp) {
			i, err := args.AsInt()
			if err != nil {
				return false, err
			}
			t.params = append(t.params, i)
			return false, nil
		}
		if t.isCompatibleType(float64Type, exp) {
			f, err := args.AsFloat64()
			if err != nil {
				return false, err
			}
			t.params = append(t.params, f)
			return false, nil
		}
		break
	}

	if exp == stringType {
		t.params = append(t.params, args.AsString())
		return true, nil
	}
	return false, fmt.Errorf("invalid type of argument: %s", args.AsString())
}

func (t *whereBuilder) VisitArgument(argumentCtx fq.ArgumentContext) {
	if t.lastSelector == nil {
		return
	}
	s, err := t.negotiateArgumentType(&argumentCtx)
	if err != nil {
		t.lastSelector = nil
		t.errors = append(t.errors, fmt.Errorf("invalid type of argument: %s", argumentCtx.AsString()))
		return
	}

	if s && (argumentCtx.StartsWithWildcard() || argumentCtx.EndsWithWildcard()) {
		t.sb.WriteString("CONCAT(")
	}
	if s && argumentCtx.StartsWithWildcard() {
		t.sb.WriteString("'%',")
	}
	parameterBuilder(t.paramStyle, len(t.params), &t.sb)
	if s && argumentCtx.EndsWithWildcard() {
		t.sb.WriteString(",'%'")
	}
	if s && (argumentCtx.StartsWithWildcard() || argumentCtx.EndsWithWildcard()) {
		t.sb.WriteString(")")
	}

}

func (a *Adapter) Where(query string) (*WherePredicate, error) {
	ast, err := a.parser.Parse(query)
	if err != nil {
		return nil, err
	}
	wb := whereBuilder{
		fields:     a.fields,
		params:     make([]interface{}, 0),
		errors:     make([]error, 0),
		delim:      a.delim,
		paramStyle: a.paramStyle,
	}
	ast.Accept(&wb)
	if len(wb.errors) > 0 {
		return nil, concatErrrors(wb.errors)
	}
	return &WherePredicate{
		sql:    wb.sb.String(),
		params: wb.params,
	}, nil
}

func (a *Adapter) OrderBy(query string) (*OrderByClause, error) {
	if query == "" {
		return &OrderByClause{}, nil
	}
	var sb strings.Builder
	s := strings.Split(query, ";")
	for i, v := range s {
		if len(v) > 1 {
			if f, ok := a.fields[strings.ToLower(v[1:])]; ok {
				delimitBuilder(a.delim, f.Db, &sb)
				if v[0] == '+' {
					sb.WriteString(" ASC")
					if i != len(s)-1 {
						sb.WriteString(", ")
					}
					continue
				} else if v[0] == '-' {
					sb.WriteString(" DESC")
					if i != len(s)-1 {
						sb.WriteString(", ")
					}
					continue
				}
			}
		}
		return nil, errors.New("invalid order by selector")
	}
	return &OrderByClause{sql: sb.String()}, nil
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

func NewAdapter(mapping FieldMapping, options ...func(*Adapter)) *Adapter {
	adapter := &Adapter{fields: mapping, parser: fq.NewParser(), paramStyle: StandardParamStyle, delim: StandardSqlDelimiter}
	for _, o := range options {
		o(adapter)
	}
	return adapter
}

func NewAdapterFor(typeDef interface{}, options ...func(*Adapter)) *Adapter {
	mapping := tagsFromStruct(typeDef)
	adapter := &Adapter{fields: mapping, parser: fq.NewParser(), paramStyle: StandardParamStyle, delim: StandardSqlDelimiter}
	for _, o := range options {
		o(adapter)
	}
	return adapter
}

const tagdef = "fiql"

func tagsFromStruct(s interface{}) FieldMapping {
	p := reflect.ValueOf(s)
	v := reflect.Indirect(p)
	if v.Kind() != reflect.Struct {
		return map[string]Field{}
	}
	m := make(map[string]Field, 0)
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		tag := f.Tag.Get(tagdef)
		if tag == "" || tag == "-" {
			continue
		}
		alias := tag
		//peek db tag from sqlx
		// db := f.Tag.Get("db") // okay probably shouldnt touch those for now ...
		db := f.Name
		if strings.Contains(tag, ",db:") {
			s := strings.Split(tag, ",db:")
			alias = s[0]
			db = s[1]
		}
		alias = strings.ToLower(alias)
		m[alias] = Field{
			Alias: alias,
			Type:  f.Type,
			Db:    db,
		}
	}
	return m
}

type Field struct {
	Db    string
	Alias string
	Type  reflect.Type
}

type WherePredicate struct {
	sql    string
	params []interface{}
}

// to satisfy sqlizer https://pkg.go.dev/github.com/masterminds/squirrel#Sqlizer
func (w *WherePredicate) ToSql() (string, []interface{}, error) {
	return w.sql, w.params, nil
}

type OrderByClause struct {
	sql string
}

func (o *OrderByClause) ToSql() (string, []interface{}, error) {
	return o.sql, nil, nil
}

func (o *OrderByClause) String() string {
	return o.sql
}

type MappingBuilder struct {
	fm FieldMapping
}

func NewMappingBuilder() *MappingBuilder {
	return &MappingBuilder{
		fm: make(FieldMapping),
	}
}

func (b *MappingBuilder) AddStringMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  stringType,
	}
	return b
}

func (b *MappingBuilder) AddDateMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  timeType,
	}
	return b
}

func (b *MappingBuilder) AddFloatMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  float64Type,
	}
	return b
}

func (b *MappingBuilder) AddIntMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  intType,
	}
	return b
}

func (b *MappingBuilder) Build() FieldMapping {
	return b.fm
}
