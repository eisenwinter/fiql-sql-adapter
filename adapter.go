package fiqlsqladapter

import (
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

type AggregateError struct {
	errors []error
	msg    string
}

func (a *AggregateError) Error() string {
	return a.msg
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
func (t *whereBuilder) VisitOperator(operator fq.OperatorDefintion) {
	switch operator {
	case fq.OperatorAND:
		t.sb.WriteString(" AND ")
	case fq.OperatorOR:
		t.sb.WriteString(" OR ")
	}
}

func (t *whereBuilder) VisitSelector(selector string) {
	if fi, ok := t.fields[strings.ToLower(selector)]; ok {
		//validate selector if viable for this query
		switch t.delim {
		case StandardSqlDelimiter:
			t.sb.WriteString(`"`)
		case MssqlDelimiter:
			t.sb.WriteString("[")
		case MariaDelimiter:
			t.sb.WriteString("`")
		}
		t.sb.WriteString(fi.Db)
		switch t.delim {
		case StandardSqlDelimiter:
			t.sb.WriteString(`"`)
		case MssqlDelimiter:
			t.sb.WriteString("]")
		case MariaDelimiter:
			t.sb.WriteString("`")
		}
		t.lastSelector = &fi
	} else {
		t.errors = append(t.errors, fmt.Errorf("invalid selector: %s", selector))
		t.lastSelector = nil
	}

}
func (t *whereBuilder) VisitComparison(comparison fq.ComparisonDefintion) {
	if t.lastSelector == nil {
		return
	}
	switch comparison {
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

func (t *whereBuilder) VisitArgument(argument string, valueCtx fq.ValueContext) {
	if t.lastSelector == nil {
		return
	}
	exp := t.lastSelector.Type
	fmt.Printf("needs to be of type %v", exp)
	s := false
	//validate if argument matches type in db
	//use t.lastSelector for this
	switch valueCtx.ValueRecommendation() {
	case fq.ValueRecommendationDateTime:
		if t.isCompatibleType(timeType, exp) {
			time, err := valueCtx.AsTime()
			if err != nil {
				t.errors = append(t.errors, fmt.Errorf("could not convert type: %w", err))
				t.lastSelector = nil
				return
			}
			t.params = append(t.params, time)
		} else {
			t.errors = append(t.errors, fmt.Errorf("invalid type of argument: %s", argument))
			return
		}
	case fq.ValueRecommendationNumber:
		if t.isCompatibleType(intType, exp) {
			number, err := valueCtx.AsInt()
			if err != nil {
				t.errors = append(t.errors, fmt.Errorf("could not convert type: %w", err))
				t.lastSelector = nil
				return
			}
			t.params = append(t.params, number)
		} else if t.isCompatibleType(float64Type, exp) {
			number, err := valueCtx.AsFloat64()
			if err != nil {
				t.errors = append(t.errors, fmt.Errorf("could not convert type: %w", err))
				t.lastSelector = nil
				return
			}
			t.params = append(t.params, number)
		} else {
			t.errors = append(t.errors, fmt.Errorf("invalid type of argument: %s", argument))
			return
		}
	case fq.ValueRecommendationDuration:
		if t.isCompatibleType(timeType, exp) {
			time, err := valueCtx.AsDuration()
			if err != nil {
				t.errors = append(t.errors, fmt.Errorf("could not convert type: %w", err))
				t.lastSelector = nil
				return
			}
			t.params = append(t.params, time)
		} else {
			t.errors = append(t.errors, fmt.Errorf("invalid type of argument: %s", argument))
			return
		}
	default:
		t.params = append(t.params, valueCtx.AsString())
		s = true
	}

	if s && (valueCtx.StartsWithWildcard() || valueCtx.EndsWithWildcard()) {
		t.sb.WriteString("CONCAT(")
	}
	if s && valueCtx.StartsWithWildcard() {
		t.sb.WriteString("'%',")
	}
	switch t.paramStyle {
	case PgParamStyle:
		t.sb.WriteString("$")
		t.sb.WriteString(strconv.Itoa(len(t.params)))
	case MssqlParamStyle:
		t.sb.WriteString("@")
		t.sb.WriteString(strconv.Itoa(len(t.params)))
	case StandardParamStyle:
		t.sb.WriteString("?")
	}
	if s && valueCtx.EndsWithWildcard() {
		t.sb.WriteString(",'%'")
	}
	if s && (valueCtx.StartsWithWildcard() || valueCtx.EndsWithWildcard()) {
		t.sb.WriteString(")")
	}

}

func (a *Adapter) Map(query string) (*WherePredicate, error) {
	//ew
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
		return nil, fmt.Errorf("error occurred: %+v", wb.errors)
	}
	return &WherePredicate{
		sql:    wb.sb.String(),
		params: wb.params,
	}, nil
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
