package fiqlsqladapter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type tagsTestOnlyStruct struct {
	Test string `fiql:"test"`
	blah string `fiql:"-"`
	blub string
}

type tagsTimeOnlyStruct struct {
	test time.Time `fiql:"time"`
	blah string    `fiql:"-"`
	blub string
}

type withDbtagStruct struct {
	test time.Time `fiql:"time,db:figgety"`
	blah string    `fiql:"-"`
	blub string
}

type tagsStringPointerStruct struct {
	Test *string `fiql:"test"`
	blah string  `fiql:"-"`
	blub string
}

type withDbtagAndTablePrefixStruct struct {
	test time.Time `fiql:"time,db:mytable.figgety"`
	blah string    `fiql:"-"`
	blub string
}

type withDbtagAndTableDoublePrefixStruct struct {
	test time.Time `fiql:"time,db:mytable.figgety.flop"`
	blah string    `fiql:"-"`
	blub string
}

func TestTagsFromStruct(t *testing.T) {
	tags := tagsFromStruct(tagsTestOnlyStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"test": Field{Db: "Test", Alias: "test", Type: stringType, TablePrefix: ""}}), tags)
}

func TestTagsFromTimeStruct(t *testing.T) {
	tags := tagsFromStruct(tagsTimeOnlyStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"time": Field{Db: "test", Alias: "time", Type: timeType, TablePrefix: ""}}), tags)
}

func TestDbTagsFromTimeStruct(t *testing.T) {
	tags := tagsFromStruct(withDbtagStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"time": Field{Db: "figgety", Alias: "time", Type: timeType, TablePrefix: ""}}), tags)
}

func TestTagsFromPtrToStruct(t *testing.T) {
	tags := tagsFromStruct(&tagsTestOnlyStruct{})
	assert.Equal(t, FieldMapping{"test": Field{Db: "Test", Alias: "test", Type: stringType, TablePrefix: ""}}, tags)
}

func TestTagsFromPtrToStructWithStringPTr(t *testing.T) {
	tags := tagsFromStruct(&tagsStringPointerStruct{})
	assert.Equal(t, FieldMapping{"test": Field{Db: "Test", Alias: "test", Type: stringPtrType, TablePrefix: ""}}, tags)
}

func TestDbTagsAndTablePrefixFromTimeStruct(t *testing.T) {
	tags := tagsFromStruct(withDbtagAndTablePrefixStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"time": Field{Db: "figgety", Alias: "time", Type: timeType, TablePrefix: "mytable"}}), tags)
}

func TestDbTagsAndDoubleTablePrefixFromTimeStruct(t *testing.T) {
	tags := tagsFromStruct(withDbtagAndTableDoublePrefixStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"time": Field{Db: "figgety.flop", Alias: "time", Type: timeType, TablePrefix: "mytable"}}), tags)
}
