package fiqlsqladapter

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

// String simply returns the query string
func (w *WherePredicate) String() string {
	return w.sql
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

// String simply returns the query string
func (o *OrderByClause) String() string {
	return o.sql
}
