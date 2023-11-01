package contesta

import (
	"fmt"
	"reflect"
)

type Failure interface {
	Failure() string
}

type UnexpectedTypeFailure struct {
	actual   reflect.Type
	expected reflect.Type
}

func (ute UnexpectedTypeFailure) Failure() string {
	return fmt.Sprintf("expected a %s but got a %s", ute.expected, ute.actual)
}

type UnexpectedKindFailure struct {
	actual   reflect.Kind
	expected reflect.Kind
}

func (ute UnexpectedKindFailure) Failure() string {
	return fmt.Sprintf("expected a %s but got a %s", ute.expected, ute.actual)
}

type NotEqualFailure struct {
	actual   any
	expected any
}

func (nef NotEqualFailure) Failure() string {
	return fmt.Sprintf("%v is not equal to %v", nef.actual, nef.expected)
}
