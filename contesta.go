package contesta

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/houseabsolute/contesta/internal/ansi"
	"github.com/jedib0t/go-pretty/v6/table"
)

// Is(
//     t,
//     val,
//     Map(
//        Key("foo").Is(42),
//        Key("bar").Is(
//            Array(
//                Next().Is("hi"),
//                Next().Is("ho"),
//            )
//        ),
//        NonExhaustive(),
//     )
// )

// C contains state for the current set of tests. You should create a new `C`
// in every `Test*` function or subtest.
type C struct {
	t                 TestingT
	callerPackageRoot string
	state             *state
	output            StringWriter
}

type state struct {
	output []outputItem
	actual []any
	paths  []Path
	caller *string
}

type outputItem struct {
	result  *result
	warning string
}

// TestingT is an interface wrapper around `*testing.T` for the portion of its
// API that we care about.
type TestingT interface {
	Fail()
	Fatal(args ...interface{})
	Helper()
}

// StringWriter is an interface used for writing strings.
type StringWriter interface {
	WriteString(string) (int, error)
}

// Contester is the interface for anything that implements the `Test` method.
type Contester interface {
	Test(c *C, value any) []*result
}

// New takes any implementer of the `TestingT` interface and returns a new
// `*contesta.C`. A `*C` created this way will send its output to `os.Stdout`.
func New(t TestingT) *C {
	return &C{
		t: t,
		// On Windows the root will be something with backslashes (C:\foo\bar)
		// but Go package paths have forward slashes (C:/foo/bar) so we
		// convert the root to the forward slash version.
		callerPackageRoot: filepath.ToSlash(filepath.Dir(findFrame(1).File)),
		output:            os.Stdout,
	}
}

// NewWithOutput takes any implementer of the `TestingT` interface and a
// `StringWriter` implementer and returns a new `*contesta.C`. This is
// provided primarily for the benefit of testing code that wants to capture
// the output from contesta.
func NewWithOutput(t TestingT, o StringWriter) *C {
	return &C{t: t, output: o}
}

// Is tests that two variables are exactly equal. The first variable is the
// actual variable and the second is what is expected. The `expect` argument
// can be either a literal value or anything that implements the
// `contesta.Contester` interface.
//
// The final arguments are the assertion name. If you provide a single
// argument, this should be a string naming the assertion. If you provide more
// than one argument, they will be formatted using `fmt.Sprintf(args[0],
// args[1]...)`. If you do not provide a name then one will be generated.
//
// Under the hood this is implemented using an ExactEqualityTester.
func (c *C) Is(actual, expected any, args ...any) bool {
	c.t.Helper()
	c.ResetState()

	actualType := reflect.TypeOf(actual)
	c.PushPath(c.NewPath(describeType(actualType), 0, "contesta.(*C).Is"))
	defer c.PopPath()

	return c.processResults(c.is(actual, expected), "Is", args)
}

// ValueIs tests that two variables contain the same value. The first variable
// is the actual variable and the second is what is expected.
//
// The final arguments follow the same rules as `c.Is`.
//
// If the two variables to be compared are of different types this is fine as
// long as one type can be converted to the other (for example `int32` and
// `int64`).
//
// Under the hood this is implemented using a `ValueEqualityTester`.
func (c *C) ValueIs(actual, expect any, args ...any) bool {
	c.t.Helper()
	c.ResetState()

	actualType := reflect.TypeOf(actual)
	c.PushPath(c.NewPath(describeType(actualType), 0, "contesta.(*C).ValueIs"))
	defer c.PopPath()

	if _, ok := expect.(Contester); ok {
		return c.processResults([]*result{{
			pass:   false,
			actual: newValue(actual),
			expect: newValue(expect),
			where:  inUsage,
			description: fmt.Sprintf(
				"You cannot pass a Contester as the expected value to ValueIs",
			),
		}}, "ValueIs", args)
	}

	vet := &ValueEqualityTester{expect}
	return c.processResults(vet.Test(c, actual), "ValueIs", args)
}

func (c *C) processResults(results []*result, method string, args []any) bool {
	passed := true
	for _, r := range results {
		c.output.WriteString(r.describe(argsToName(method, args), ansi.DefaultScheme))
		if !r.pass {
			c.t.Fail()
			passed = false
		}
	}

	return passed
}

func maybeNot(r *result) string {
	if r.pass {
		return "   "
	}
	return "not"
}

// ResetState resets the internal state of the `*contesta.D` struct. This is
// public for the benefit of test packages that want to provide their own
// testers or test functions like `contesta.Is`.
func (c *C) ResetState() {
	c.state = &state{}
}

func (c *C) is(actual, expected any) []*result {
	if e, ok := expected.(Contester); ok {
		return e.Test(c, actual)
	}
	eet := &ExactEqualityTester{expected}
	return eet.Test(c, actual)
}

// Paths returns the current paths, with the caller overriden by the last
// caller if one was set.
func (c *C) Paths() []Path {
	paths := c.state.paths
	if c.state.caller != nil {
		paths[len(paths)-1].caller = *c.state.caller
	}
	return paths
}

// PushPath adds a path to the current path stack.
func (c *C) PushPath(path Path) {
	c.state.paths = append(c.state.paths, path)
}

// PopPath removes the top path from the current path stack.
func (c *C) PopPath() {
	if len(c.state.paths) > 0 {
		c.state.paths = c.state.paths[:len(c.state.paths)-1]
	}
}

// SetCaller adds a caller that overrides the current path.
func (c *C) SetCaller(caller string) {
	c.state.caller = &caller
}

// PopPath removes the top path from the current path stack.
func (c *C) UnsetCaller() {
	c.state.caller = nil
}

func argsToName(defaultName string, args []any) string {
	if len(args) == 0 {
		return defaultName
	}

	format, ok := args[0].(string)
	if !ok {
		format = fmt.Sprintf("%v", args[0])
	}

	if len(args) > 1 {
		return fmt.Sprintf(format, args[1:]...)
	}

	return format
}

func (c *C) ok(results []*result, name string) bool {
	pass, err := c.renderOutput(results, name)
	if err != nil {
		panic(err)
	}

	return pass
}

func (c *C) renderOutput(results []*result, name string) (bool, error) {
	pass := true
	scheme := ansi.DefaultScheme

	var warnings []string
	for _, r := range results {
		// nolint: gocritic
		if r.pass {
			_, err := c.output.WriteString(fmt.Sprintf("Assertion ok: %s\n", name))
			if err != nil {
				return false, err
			}
		} else {
			pass = false
			c.t.Fail()
			_, err := c.output.WriteString(r.describe(name, scheme))
			if err != nil {
				return pass, err
			}
		}

		// XXX where do warnings go?
		// } else if o.warning != "" {
		// 		warnings = append(warnings, o.warning)
		// 	} else {
		// 		return pass, errors.New("we have an output which does not have a result or a warning but that should never happen")
		// 	}
	}

	if len(warnings) != 0 {
		var title string
		if len(warnings) == 1 {
			title = "Warning"
		} else {
			title = "Warnings"
		}
		tw := tableWithTitle(title, scheme)
		for _, w := range warnings {
			tw.AppendRow(table.Row{scheme.Warning(w)})
		}
		_, err := c.output.WriteString(tw.Render() + "\n")
		if err != nil {
			return pass, err
		}
	}

	if len(warnings) > 0 || !pass {
		// Needed to separate a table + warnings from the next batch.
		_, err := c.output.WriteString("\n")
		if err != nil {
			return pass, err
		}
	}

	return pass, nil
}

// CalledAt returns a string describing the function, file, and line for this
// path element.
func (p Path) CalledAt() string {
	return fmt.Sprintf("%s called %s", p.caller, p.callee)
}
