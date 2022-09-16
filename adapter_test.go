package fiqlsqladapter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasicSqlAdapter(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m)
	res, err := adapter.Map("ml==life")
	assert.NoError(t, err)
	sql, params, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, `("mylife" LIKE ?)`, sql)
	assert.Equal(t, []interface{}{"life"}, params)
}

func TestBasicSqlAndAdapter(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m)
	res, err := adapter.Map("ml==life;ml==hard")
	assert.NoError(t, err)
	sql, params, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, `("mylife" LIKE ? AND "mylife" LIKE ?)`, sql)
	assert.Equal(t, []interface{}{"life", "hard"}, params)
}

func TestBasicSqlOrAdapter(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m)
	res, err := adapter.Map("ml==life,ml==hard")
	assert.NoError(t, err)
	sql, params, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, `("mylife" LIKE ? OR "mylife" LIKE ?)`, sql)
	assert.Equal(t, []interface{}{"life", "hard"}, params)
}

func TestBasicSqlNestedAdapter(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	m["lo"] = Field{
		Db:    "love",
		Alias: "lo",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m)
	res, err := adapter.Map("(ml==life;lo==me),(ml==hard;lo==you)")
	assert.NoError(t, err)
	sql, params, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, `(("mylife" LIKE ? AND "love" LIKE ?) OR ("mylife" LIKE ? AND "love" LIKE ?))`, sql)
	assert.Equal(t, []interface{}{"life", "me", "hard", "you"}, params)
}

func TestBasicSqlNestedAdapterMSSQL(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	m["lo"] = Field{
		Db:    "love",
		Alias: "lo",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m, WithDialectMSSQL())
	res, err := adapter.Map("(ml==life;lo==me),(ml==hard;lo==you)")
	assert.NoError(t, err)
	sql, params, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, `(([mylife] LIKE @1 AND [love] LIKE @2) OR ([mylife] LIKE @3 AND [love] LIKE @4))`, sql)
	assert.Equal(t, []interface{}{"life", "me", "hard", "you"}, params)
}

func TestBasicSqlNestedAdapterPostGres(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	m["lo"] = Field{
		Db:    "love",
		Alias: "lo",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m, WithDialectPostgres())
	res, err := adapter.Map("(ml==life;lo==me),(ml==hard;lo==you)")
	assert.NoError(t, err)
	sql, params, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, `(("mylife" LIKE $1 AND "love" LIKE $2) OR ("mylife" LIKE $3 AND "love" LIKE $4))`, sql)
	assert.Equal(t, []interface{}{"life", "me", "hard", "you"}, params)
}

func TestBasicWildCardLeading(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m, WithDialectMSSQL())
	res, err := adapter.Map("ml==*life")
	assert.NoError(t, err)
	sql, _, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "([mylife] LIKE CONCAT('%',@1))", sql)
}

func TestBasicWildCardTrailing(t *testing.T) {
	m := make(FieldMapping)
	m["ml"] = Field{
		Db:    "mylife",
		Alias: "ml",
		Type:  reflect.TypeOf(""),
	}
	adapter := NewAdapter(m, WithDialectMSSQL())
	res, err := adapter.Map("ml==life*")
	assert.NoError(t, err)
	sql, _, err := res.ToSql()
	assert.NoError(t, err)
	assert.Equal(t, "([mylife] LIKE CONCAT(@1,'%'))", sql)
}

type myFunnyRowStruct struct {
	ID       int        `fiql:"id"`
	Fee      *float64   `fiql:"fee,db:fee"`
	Amount   float64    `fiql:"amt,db:amount"`
	Currency *string    `fiql:"cur" db:"currency"`
	Tx       string     `fiql:"tx"`
	Created  time.Time  `fiql:"cre,db:created_at"`
	Updated  *time.Time `fiql:"upd,db:updated_at"`
	Secret   string     //equals -
	Blocked  string     `fiql:"-"`
}

func TestFromStructBasic(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	res, err := adp.Map("id==1")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	s, _, err := res.ToSql()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, `("ID" = $1)`, s)
}

func TestFromStructBasicMultiple(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	res, err := adp.Map("id==1;(cre=lt=-P1D,upd=gt=2022-09-16T10:15:04Z)")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	s, _, err := res.ToSql()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, `("ID" = $1 AND ("created_at" < $2 OR "updated_at" > $3))`, s)
}

func TestFromStructFloatsAndInts(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	res, err := adp.Map("id==1;fee=le=0.0;amt=gt=0")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	s, args, err := res.ToSql()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, `("ID" = $1 AND "fee" <= $2 AND "amount" > $3)`, s)
	assert.Equal(t, []interface{}([]interface{}{1, 0.0, 0.0}), args)
}

func TestFromStructSecret(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	_, err := adp.Map("secret==1")
	assert.Error(t, err)
}

func TestFromStructBlocked(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	_, err := adp.Map("blocked==1")
	assert.Error(t, err)
}

func TestFromStructWildCardLeading(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	res, err := adp.Map("tx==*001020")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	s, _, err := res.ToSql()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, `("Tx" LIKE CONCAT('%',$1))`, s)
}

func TestFromStructWildCardTrailing(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	res, err := adp.Map("tx==001020*")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	s, _, err := res.ToSql()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, `("Tx" LIKE CONCAT($1,'%'))`, s)
}

func TestFromStructWildCard(t *testing.T) {
	adp := NewAdapterFor(&myFunnyRowStruct{}, WithDialectPostgres())
	res, err := adp.Map("tx==*001020*")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	s, _, err := res.ToSql()
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, `("Tx" LIKE CONCAT('%',$1,'%'))`, s)
}
