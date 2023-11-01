package contesta

import (
	"fmt"
	"reflect"
)

type MapTester struct {
	tests  []MapTest
	caller string
}

type MapTest interface {
	isMapTest()
	Test(c *C, value reflect.Value) []*result
}

func (c *C) Map(test MapTest, tests ...MapTest) *MapTester {
	tests = append([]MapTest{test}, tests...)

	// var hasEnd bool
	// for _, t := range tests {
	// 	if _, ok := t.(ExhaustiveContainerTest); ok {
	// 		hasEnd = true
	// 		break
	// 	}
	// 	if _, ok := t.(NonExhaustiveContainerTest); ok {
	// 		hasEnd = true
	// 		break
	// 	}
	// }
	// if !hasEnd {
	// 	tests = append(tests, End())
	// }

	return &MapTester{
		tests:  tests,
		caller: c.Caller(),
	}
}

func (mt *MapTester) Test(c *C, actual any) []*result {
	c.SetCaller(mt.caller)
	defer c.UnsetCaller()

	vt := reflect.TypeOf(actual)
	if vt.Kind() != reflect.Map {
		actualKind := vt.Kind()
		return []*result{{
			pass: false,
			description: fmt.Sprintf(
				"Expected a map but got %s %s",
				Article(actualKind.String()),
				actualKind,
			),
			paths: c.Paths(),
			where: inType,
		}}
	}

	va := reflect.ValueOf(actual)

	var res []*result
	for _, t := range mt.tests {
		res = append(res, t.Test(c, va)...)
	}

	return res
}

type IncompleteMapKeyTest struct {
	key any
	c   *C
}

type MapKeyTest struct {
	key    any
	test   any
	caller string
}

func (c *C) Key(key any) IncompleteMapKeyTest {
	return IncompleteMapKeyTest{key, c}
}

func (imt IncompleteMapKeyTest) Is(expected any) MapKeyTest {
	return MapKeyTest{
		key:    imt.key,
		test:   expected,
		caller: imt.c.Caller(),
	}
}

func (MapKeyTest) isMapTest() {}

func (mt MapKeyTest) Test(c *C, value reflect.Value) []*result {
	return c.is(value.MapIndex(reflect.ValueOf(mt.key)).Interface(), mt.test)
}

type MapKeysNotCheckedFailure struct {
	keys []any
}

func (mknf MapKeysNotCheckedFailure) Failure() string {
	return fmt.Sprintf("%d keys in the map were not checked: %v", len(mknf.keys), mknf.keys)
}
