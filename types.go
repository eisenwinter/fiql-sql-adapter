package fiqlsqladapter

import "fmt"

// WherePredicate holds the content of a where predicate
// it will always be wrapped in outter most braces to avoid
// issues with OR chaining leading to unwated results
// it does not contain the WHERE keyword
type WherePredicate struct {
	sql    string
	params []interface{}
}

// ToSql returns the query string and the parameters
// satisfies sqlizer https://pkg.go.dev/github.com/masterminds/squirrel#Sqlizer
func (w *WherePredicate) ToSql() (string, []interface{}, error) {
	return w.sql, w.params, nil
}

// Query satisfies querier interface from ent
// https://pkg.go.dev/entgo.io/ent@v0.12.4/dialect/sql#Querier
func (w *WherePredicate) Query() (string, []any) {
	return w.sql, w.params
}

// Sql returns the underlying sql string
func (w *WherePredicate) Sql() string {
	return w.sql
}

// Parameters return the underlying parameters
func (w *WherePredicate) Parameters() []interface{} {
	return w.params
}

// String simply returns the query string and paremeters
func (w *WherePredicate) String() string {
	return fmt.Sprintf("%s (%+v)", w.sql, w.params)
}

// OrderByClause represents a order by clause
// without the order by keyword
type OrderByClause struct {
	sql string
}

// ToSql returns the query string and the parameters
// satisfies sqlizer https://pkg.go.dev/github.com/masterminds/squirrel#Sqlizer
func (o *OrderByClause) ToSql() (string, []interface{}, error) {
	return o.sql, nil, nil
}

// Sql returns the underlying sql string
func (w *OrderByClause) Sql() string {
	return w.sql
}

// String simply returns the query string
func (o *OrderByClause) String() string {
	return o.sql
}
