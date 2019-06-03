package nulls

import (
	"database/sql/driver"
	"reflect"

	"github.com/gofrs/uuid"
)

// nullable a generic representation of nulls type.
type nullable interface {
	Interface() interface{}
	Value() (driver.Value, error)
}

// Nulls a generic nulls type. something that implements
// nullable interface. can be any of nulls.Int, nulls.uuid.UUID
// nulls.String, etc.
type Nulls struct {
	Value interface{}
}

// Interface calls Interface function for value.
func (nulls *Nulls) Interface() interface{} {
	n := nulls.Value.(nullable)
	return n.Interface()
}

// WrappedValue returns the wrapped value for a nulls
// implementation.
func (nulls *Nulls) WrappedValue() interface{} {
	v := reflect.ValueOf(nulls.Value)
	switch {
	case v.FieldByName("Int").IsValid():
		return v.FieldByName("Int").Interface()
	case v.FieldByName("Int64").IsValid():
		return v.FieldByName("Int64").Interface()
	case v.FieldByName("UUID").IsValid():
		return v.FieldByName("UUID").Interface()
	default:
		return nil
	}
}

// Parse parses the specified value to the corresponding
// nullable type. value is one of the inner value hold
// by a nullable type. i.e int, string, uuid.UUID etc.
func (nulls *Nulls) Parse(value interface{}) interface{} {
	switch nulls.Value.(type) {
	case Int:
		return NewInt(value.(int))
	case Int64:
		return NewInt64(value.(int64))
	case UUID:
		return NewUUID(value.(uuid.UUID))
	default:
		return value
	}
}

// New returns a wrapper called nulls for the
// interface passed as a param.
func New(i interface{}) *Nulls {
	if _, ok := i.(nullable); !ok {
		return nil
	}
	return &Nulls{Value: i}
}
