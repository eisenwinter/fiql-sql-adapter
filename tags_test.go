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

func TestTagsFromStruct(t *testing.T) {
	tags := tagsFromStruct(tagsTestOnlyStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"test": Field{Db: "Test", Alias: "test", Type: stringType}}), tags)
}

func TestTagsFromTimeStruct(t *testing.T) {
	tags := tagsFromStruct(tagsTimeOnlyStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"time": Field{Db: "test", Alias: "time", Type: timeType}}), tags)
}

func TestDbTagsFromTimeStruct(t *testing.T) {
	tags := tagsFromStruct(withDbtagStruct{})
	assert.Equal(t, FieldMapping(FieldMapping{"time": Field{Db: "figgety", Alias: "time", Type: timeType}}), tags)
}

func TestTagsFromPtrToStruct(t *testing.T) {
	tags := tagsFromStruct(&tagsTestOnlyStruct{})
	assert.Equal(t, FieldMapping{"test": Field{Db: "Test", Alias: "test", Type: stringType}}, tags)
}
