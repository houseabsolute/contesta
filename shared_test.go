package contesta

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ContestaRecorder struct {
	*C
	record []*state
}

func NewRecorder(c *C) *ContestaRecorder {
	return &ContestaRecorder{
		c,
		[]*state{},
	}
}

// func (cr *ContestaRecorder) Is(actual, expect interface{}, args ...interface{}) bool {
// 	ok := cr.C.Is(actual, expect, args...)
// 	cr.record = append(cr.record, cr.c.state)
// 	return ok
// }

// func (cr *ContestaRecorder) Passes(actual, expect interface{}, args ...interface{}) bool {
// 	ok := cr.c.Passes(actual, expect.(Comparer), args...)
// 	cr.record = append(cr.record, cr.c.state)
// 	return ok
// }

type mockT struct {
	calls []Call
}

type Call struct {
	Method string
	Args   []interface{}
}

func newMockT() *mockT {
	return &mockT{}
}

func (mt *mockT) called(args ...interface{}) {
	pc := make([]uintptr, 1)
	n := runtime.Callers(2, pc)
	if n == 0 {
		panic("Cannot get caller from runtime.Callers!")
	}
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	mt.calls = append(mt.calls, Call{Method: methodName(frame.Function), Args: args})
}

func methodName(f string) string {
	s := strings.Split(f, ".")
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}

func (mt *mockT) AssertNotCalled(t *testing.T, method string) {
	for _, c := range mt.calls {
		if c.Method == method {
			t.Errorf("The %s method was called when it should not have been", method)
			return
		}
	}
}

func (mt *mockT) AssertCalled(t *testing.T, method string, args ...interface{}) {
	for _, c := range mt.calls {
		if c.Method == method {
			_, differences := mock.Arguments(args).Diff(c.Args)
			if differences == 0 {
				return
			}

			t.Errorf(
				"Expected the %s method to be called with:\n%v\nbut it was called with:\n%v\n",
				method,
				args,
				c.Args,
			)
			return
		}
	}
	t.Errorf("Expected the %s method to be called with:\n%v\nbut it was never called", method, args)
}

func (mt *mockT) AssertPassed(t *testing.T) {
	mt.AssertNotCalled(t, "Fail")
	mt.AssertNotCalled(t, "Fatal")
}

func (mt *mockT) AssertFailed(t *testing.T) {
	mt.AssertCalled(t, "Fail")
}

func (mt *mockT) FindCall(method string) *Call {
	for _, c := range mt.calls {
		if c.Method == method {
			c := c
			return &c
		}
	}
	return nil
}

func (mt *mockT) Fail() {
	mt.called()
}

func (mt *mockT) Fatal(args ...interface{}) {
	mt.called(args)
}

func (mt *mockT) Helper() {
	mt.called()
}

func (mt *mockT) WriteString(s string) (int, error) {
	mt.called(s)
	return len([]byte(s)), nil
}

type resultExpect struct {
	pass     bool
	dataPath []string
}

func AssertResultsAre(t *testing.T, actual []outputItem, expect []resultExpect, name string) {
	if assert.Len(t, actual, len(expect), "got %d result(s)", len(expect)) {
		for i := range expect {
			i := i
			t.Run(fmt.Sprintf("results[%d]", i), func(t *testing.T) {
				assert.Equal(t, expect[i].pass, actual[i].result.pass, "got expected pass", i)
				dataPath := []string{}
				for _, p := range actual[i].result.paths {
					dataPath = append(dataPath, p.data)
				}
				assert.Equal(t, expect[i].dataPath, dataPath, "got expected data path")
			})
		}
	}
}

// type GTComparer int

// func (sc GTComparer) Compare(c *C) {
// 	c.AddResult(result{
// 		pass:   c.Actual().(int) > int(sc),
// 		actual: newValue(c.Actual()),
// 		expect: newValue(sc),
// 		op:     ">",
// 		where:  inValue,
// 	})
// }
