package fiqlsqladapter

import (
	"reflect"
	"strings"
	"time"
)

var stringType = reflect.TypeOf("")
var timeType = reflect.TypeOf(time.Time{})

var float64Type = reflect.TypeOf(float64(0))
var intType = reflect.TypeOf(int(0))

type Field struct {
	Db    string
	Alias string
	Type  reflect.Type
}
type FieldMapping map[string]Field

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
