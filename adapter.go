package fiqlsqladapter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	fq "github.com/eisenwinter/fiql-parser"
)

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

// Adapter adapter is a fiql2sql adapter
// one adapter per table is needed unless all tables
// carry the same columns - but i would not recommend sharing
// adapters for multiple  tables
type Adapter struct {
	fields     FieldMapping
	parser     *fq.Parser
	delim      delimiterStyle
	paramStyle paramStyle
	concat     concatSupport
	tableName  string
}

type whereBuilder struct {
	sb           strings.Builder
	params       []interface{}
	errors       []error
	lastSelector *Field
	fields       FieldMapping
	delim        delimiterStyle
	paramStyle   paramStyle
	concat       concatSupport
	tableName    string
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
		if fi.TablePrefix != "" {
			delimitBuilder(t.delim, fi.TablePrefix, &t.sb)
			t.sb.WriteRune('.')
		} else if t.tableName != "" {
			delimitBuilder(t.delim, t.tableName, &t.sb)
			t.sb.WriteRune('.')
		}
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
	if args.ValueRecommendation() == fq.ValueRecommendationString && isPointerCompatibleType(exp, stringType) {
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
			duration, err := args.AsDuration()
			if err != nil {
				return false, err
			}
			// we just convert the duration to a go time
			// so we dont have to worry about any further driver issues
			// with custom types
			p := time.Now().UTC().Add(time.Duration(duration.AsMilliseconds()) * time.Millisecond)
			t.params = append(t.params, p)
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
		if t.concat == concatFunctionSupported {
			t.sb.WriteString("CONCAT(")
		}
	}
	if s && argumentCtx.StartsWithWildcard() {
		if t.concat == concatFunctionSupported {
			t.sb.WriteString("'%',")
		} else {
			t.sb.WriteString("'%' || ")
		}

	}
	parameterBuilder(t.paramStyle, len(t.params), &t.sb)
	if s && argumentCtx.EndsWithWildcard() {
		if t.concat == concatFunctionSupported {
			t.sb.WriteString(",'%'")
		} else {
			t.sb.WriteString(" || '%'")
		}

	}
	if s && (argumentCtx.StartsWithWildcard() || argumentCtx.EndsWithWildcard()) {
		if t.concat == concatFunctionSupported {
			t.sb.WriteString(")")
		}
	}

}

// Where generates a where predicate from a given fiql query
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
		concat:     a.concat,
		tableName:  a.tableName,
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

// OrderBy generates a order by clause from a given query
// this is no fiql but rather the format (+/-)ALIAS[;(+/-)ALIAS]*
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

// NewAdapter returns a new fiql adapter for the given field mapping
// use the MappingBuilder to create field mapping
func NewAdapter(mapping FieldMapping, options ...func(*Adapter)) *Adapter {
	adapter := &Adapter{fields: mapping, parser: fq.NewParser(), paramStyle: standardParamStyle, delim: standardSqlDelimiter}
	for _, o := range options {
		o(adapter)
	}
	return adapter
}

// NewAdapterFor creates a new adapter from struct tags of the typeDef argument
func NewAdapterFor(typeDef interface{}, options ...func(*Adapter)) *Adapter {
	mapping := tagsFromStruct(typeDef)
	adapter := &Adapter{fields: mapping, parser: fq.NewParser(), paramStyle: standardParamStyle, delim: standardSqlDelimiter}
	for _, o := range options {
		o(adapter)
	}
	return adapter
}

// WithTableName creates an adapter that prefixes all columns with the given table name
func WithTableName(tableName string) func(*Adapter) {
	return func(a *Adapter) {
		a.tableName = tableName
	}
}
