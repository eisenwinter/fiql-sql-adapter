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

// Field is a fiql field to database column mapping
type Field struct {
	Db    string
	Alias string
	Type  reflect.Type
}

// FieldMapping is a table mapping of fields
type FieldMapping map[string]Field

// MappingBuilder helps build FieldMappings manually if no struct tags are used
type MappingBuilder struct {
	fm FieldMapping
}

// NewMappingBuilder returns a new instance of a MappingBuilder
func NewMappingBuilder() *MappingBuilder {
	return &MappingBuilder{
		fm: make(FieldMapping),
	}
}

// AddStringMapping adds a column to fiql selector mapping for a string column
func (b *MappingBuilder) AddStringMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  stringType,
	}
	return b
}

// AddDateMapping adds a column to fiql selector mapping for a date(time) column
func (b *MappingBuilder) AddDateMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  timeType,
	}
	return b
}

// AddFloatMapping adds a column to fiql selector mapping for a decimal column
func (b *MappingBuilder) AddFloatMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  float64Type,
	}
	return b
}

// AddIntMapping adds a column to fiql selector mapping for a numeric column
func (b *MappingBuilder) AddIntMapping(column, selector string) *MappingBuilder {
	b.fm[strings.ToLower(selector)] = Field{
		Alias: selector,
		Db:    column,
		Type:  intType,
	}
	return b
}

// Build returns the generated
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

		parts := strings.Split(tag, ",")
		alias := parts[0]
		db := f.Name

		if len(parts) > 1 {
			for _, v := range parts[1:] {
				if strings.HasPrefix(v, "db:") {
					db = strings.TrimPrefix(v, "db:")
				}
			}
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
