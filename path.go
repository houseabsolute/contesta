package contesta

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Path is used to track the data Path as a test goes through a complex data
// structure. It records a place in a data structure along with information
// about the call stack at that particular point in the data Path.
type Path struct {
	data   string
	callee string
	caller string
}

var ourPackages = map[string]bool{}
var stdlibRoot string

// nolint: gochecknoinits
func init() {
	ourPackages[packageFromFrame(findFrame(0))] = true

	var f runtime.Frame
	// We need to call a func in the stdlib that calls a func in our package
	// so we can use the call stack to find the stdlib root.
	strings.IndexFunc("x", func(r rune) bool {
		f = findFrame(2)
		return true
	})
	stdlibRoot = filepath.Dir(filepath.Dir(f.File))
}

// RegisterPackage adds the caller's package to the list of "internal"
// packages for the purposes of presenting paths in test failure
// output. Specifically, when a function in a registered package is found as
// the caller for a path, contesta will use the function name as the caller
// rather than showing the file and line where the call occurred.
func RegisterPackage() {
	ourPackages[packageFromFrame(findFrame(1))] = true
}

func findFrame(s int) runtime.Frame {
	pc := make([]uintptr, 1)
	n := runtime.Callers(s+1, pc)
	if n == 0 {
		panic("Cannot get New() from runtime.Callers!")
	}
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	return frame
}

var packageRE = regexp.MustCompile(`((?:[^/]+/)*[^\.]+)\.`)

func packageFromFrame(frame runtime.Frame) string {
	m := packageRE.FindStringSubmatch(frame.Function)
	if len(m) <= 1 {
		return ""
	}
	return m[1]
}

var funcNameRE = regexp.MustCompile(`^.+/`)

// NewPath takes a data path element, the number of frames to skip, and an
// optional function name. It returns a new `path` struct. If the function
// name is given, then this is used as the called function rather than looking
// at the call frames .
//
// When the desired frame is from a package marked as internal to contesta,
// then the caller's line and file is replaced with a function name so that we
// don't show (unhelpful) information about the contesta internals when
// displaying the path.
func (c *C) NewPath(data string, skip int, function string) Path {
	pc := make([]uintptr, 2)
	// The hard-coded "2" is here because we want to skip this frame and the
	// frame of the caller. We're interested in the frames before that.
	n := runtime.Callers(2+skip, pc)
	if n == 0 {
		return Path{data: data}
	}

	frames := runtime.CallersFrames(pc)
	frame, more := frames.Next()
	callee := calleeFromFrame(frame, function)

	if !more {
		return Path{
			data:   data,
			callee: funcNameRE.ReplaceAllLiteralString(callee, ""),
		}
	}

	frame, _ = frames.Next()

	return Path{
		data:   data,
		callee: funcNameRE.ReplaceAllLiteralString(callee, ""),
		caller: c.callerFromFrame(frame),
	}
}

func (c *C) Caller() string {
	pc := make([]uintptr, 3)
	// The hard-coded "3" is here because we want to skip this frame, the
	// frame of the caller, and the frame of the caller's caller. We're
	// interested in the frame before that.
	n := runtime.Callers(3, pc)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()
	return c.callerFromFrame(frame)
}

// calledAt returns a string describing the function, file, and line for this
// path element.
func (p Path) calledAt() string {
	return fmt.Sprintf("%s called %s", p.caller, p.callee)
}

func calleeFromFrame(frame runtime.Frame, function string) string {
	if function != "" {
		return function
	}

	callee := frame.Function
	if callee == "" {
		callee = "<unknown>"
	}

	return callee
}

func (c *C) callerFromFrame(frame runtime.Frame) string {
	if ourPackages[packageFromFrame(frame)] {
		return funcNameRE.ReplaceAllLiteralString(frame.Function, "")
	}

	file := frame.File
	// If the caller is in the package that created our *D then we can strip
	// that from the caller path and just show a path relative to the package
	// root.
	if strings.HasPrefix(file, c.callerPackageRoot) {
		file = strings.TrimPrefix(file, c.callerPackageRoot)[1:]
	}

	// If the caller is in the stdlib we don't need to print the entire path
	// to the stdlib.
	if strings.HasPrefix(file, stdlibRoot) {
		file = strings.Replace(file, stdlibRoot, "<stdlib>", 1)
	}

	return fmt.Sprintf("%s@%d", file, frame.Line)
}
