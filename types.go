package fiqlsqladapter

type WherePredicate struct {
	sql    string
	params []interface{}
}

// to satisfy sqlizer https://pkg.go.dev/github.com/masterminds/squirrel#Sqlizer
func (w *WherePredicate) ToSql() (string, []interface{}, error) {
	return w.sql, w.params, nil
}

func (w *WherePredicate) String() string {
	return w.sql
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
