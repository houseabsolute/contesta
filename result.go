package contesta

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/houseabsolute/contesta/internal/ansi"
	"github.com/houseabsolute/contesta/internal/term"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mattn/go-runewidth"
)

// We wrap values in a struct so that we can use a nil *value to indicate that
// the value wasn't present, as opposed to a nil *value.value, which is a nil
// _value_.
type value struct {
	value any
	desc  string
}

type result struct {
	pass        bool
	actual      *value
	expect      *value
	op          string
	paths       []Path
	caller      *string
	where       failure
	description string
}

type failure int

const (
	inType failure = iota
	inValue
	inDataStructure
	inUsage
)

func newValue(val any) *value {
	return &value{value: val}
}

func (r result) hasPaths() bool {
	return len(r.paths) != 0
}

func (r result) showActual() bool {
	return r.actual != nil
}

func (r result) showExpect() bool {
	return r.expect != nil
}

type describer struct {
	r  result
	tw table.Writer
	s  ansi.Scheme
}

func (r result) describe(name string, s ansi.Scheme) string {
	if r.pass {
		return s.Correct(s.Strong(fmt.Sprintf("Assertion ok: %s", name))) + "\n\n"
	}

	table := tableWithTitle(fmt.Sprintf("Assertion not ok: %s", name), s)
	return describer{r, table, s}.table() + "\n"
}

func tableWithTitle(title string, s ansi.Scheme) table.Writer {
	tw := table.NewWriter()
	tw.SetTitle(s.Strong(title))
	st := table.StyleDefault
	st.Box = table.StyleBoxLight
	st.Format.Header = text.FormatDefault
	st.Format.Footer = text.FormatDefault
	tw.SetStyle(st)
	return tw
}

func (d describer) table() string {
	d.addHeaders()

	footer, widths := d.footer()
	rowLen := len(footer)

	body := []table.Row{}
	for _, p := range d.r.paths {
		row := table.Row{p.data}
		w := displayWidth(p.data)
		if w > widths["PATH"] {
			widths["PATH"] = w
		}
		for i := 0; i < rowLen-2; i++ {
			row = append(row, "")
		}

		called := p.CalledAt()
		row = append(row, called)
		w = displayWidth(called)
		if w > widths["CALLER"] {
			widths["CALLER"] = w
		}

		body = append(body, row)
	}
	for _, b := range body {
		d.tw.AppendRow(b, table.RowConfig{AutoMerge: true})
	}

	d.tw.AppendFooter(footer)

	cc := columnConfigs(widths)
	if cc != nil {
		d.tw.SetColumnConfigs(cc)
	}

	var post string
	if d.r.description != "" {
		post = d.s.Strong(d.s.Incorrect(d.r.description)) + "\n"
	}

	return d.tw.Render() + "\n" + post
}

func (d describer) addHeaders() {
	header := table.Row{}
	if d.r.hasPaths() {
		header = append(header, "PATH")
	}

	if d.r.showActual() {
		header = append(header, "GOT")
	}
	if d.r.op != "" {
		header = append(header, "OP")
	}

	if d.r.showExpect() {
		header = append(header, "EXPECT")
	}
	if d.r.hasPaths() {
		header = append(header, "CALLER")
	}

	d.tw.AppendHeader(header, table.RowConfig{AutoMerge: true})
}

func (d describer) footer() ([]any, map[string]int) {
	widths := map[string]int{"PATH": 0, "CALLER": 0}

	var actual, expect, op string
	if d.r.showActual() {
		actual = fmt.Sprintf("%v", d.r.actual.value)
		widths["GOT"] = displayWidth(actual)
	}
	if d.r.showExpect() {
		expect = fmt.Sprintf("%v", d.r.expect.value)
		widths["ACTUAL"] = displayWidth(actual)
	}
	op = d.r.op

	var aType, eType string
	if d.r.showActual() {
		aType = d.r.actual.description()
		w := displayWidth(aType)
		if w > widths["GOT"] {
			widths["GOT"] = w
		}
	}
	if d.r.showExpect() {
		eType = d.r.expect.description()
		w := displayWidth(eType)
		if w > widths["ACTUAL"] {
			widths["ACTUAL"] = w
		}
	}

	switch d.r.where {
	case inType:
		aType = d.s.Incorrect(aType)
		eType = d.s.Correct(eType)
	case inValue:
		actual = d.s.Incorrect(actual)
		expect = d.s.Correct(expect)
	case inDataStructure:
		op = d.s.Incorrect(op)
	}

	footer := table.Row{}
	if d.r.hasPaths() {
		footer = append(footer, "")
	}
	if d.r.showActual() {
		footer = append(footer, d.s.Em(aType)+"\n"+actual)
	}
	if op != "" {
		// The extra space is required to make go-pretty render the right
		// border for the first line of this cell.
		footer = append(footer, " \n"+op)
		widths["OP"] = displayWidth(op)
	}
	if d.r.showExpect() {
		footer = append(footer, d.s.Em(eType)+"\n"+expect)
	}
	if d.r.hasPaths() {
		footer = append(footer, "")
	}

	return footer, widths
}

func (v *value) description() string {
	if v.desc != "" {
		return v.desc
	}

	v.desc = describeTypeOfActualValue(v.value)
	if v.value == nil {
		v.desc += " <nil>"
	}
	return v.desc
}

func describeTypeOfActualValue(val any) string {
	return describeType(reflect.TypeOf(val))
}

func describeTypeOfReflectValue(v reflect.Value) string {
	if !v.IsValid() {
		return "nil"
	}
	return describeType(v.Type())
}

func describeType(ty reflect.Type) string {
	if ty == nil {
		return "nil"
	}

	// This is only true for built-in types like string, float64, etc. If it's
	// not composite or it's not a built-in then the name doesn't match the
	// kind.
	if ty.Kind().String() == ty.Name() {
		return ty.Name()
	}

	if ty.Name() != "" {
		return ty.Name()
	}

	switch ty.Kind() {
	case reflect.Array:
		return fmt.Sprintf("[%d]", ty.Len()) + describeType(ty.Elem())
	case reflect.Chan:
		return fmt.Sprintf("chan(%s)", describeType(ty.Elem()))
	case reflect.Func:
		return describeFunc(ty)
	case reflect.Interface:
		// Can this happen?
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", describeType(ty.Key()), describeType(ty.Elem()))
	case reflect.Ptr:
		return "*" + describeType(ty.Elem())
	case reflect.Slice:
		return "[]" + describeType(ty.Elem())
	case reflect.Struct:
		return describeStruct(ty)
	case reflect.UnsafePointer:
		return "*<unsafe>"
	}

	// wtf - should not get here
	return ""
}

func describeFunc(ty reflect.Type) string {
	desc := "func "
	if name := ty.Name(); name != "" {
		desc = desc + name + " "
	}

	desc += "("
	for i := 0; i < ty.NumIn(); i++ {
		desc += describeType(ty.In(i))
	}
	if ty.IsVariadic() {
		desc += "..."
	}
	desc += ") "

	if ty.NumOut() > 1 {
		desc += "("
	}
	for i := 0; i < ty.NumOut(); i++ {
		desc += describeType(ty.Out(i))
	}
	if ty.NumOut() > 1 {
		desc += ")"
	}

	return desc
}

func describeStruct(ty reflect.Type) string {
	if ty.Name() != "" {
		return ty.Name()
	}

	return "<anon struct>"
}

func displayWidth(content string) int {
	return runewidth.StringWidth(ansi.Strip(content))
}

// XXX - if the table title is less than the calculated widths then the title
// wraps. This should be fixed.
func columnConfigs(widths map[string]int) []table.ColumnConfig {
	var total int
	for _, w := range widths {
		total += w
		// 2 for padding, 1 for separator
		total += 3
	}
	// Left most border
	total++

	w := termWidth()
	if total <= w {
		return nil
	}

	for total > w {
		diff := total - w
		if diff < 10 {
			widths["CALLER"] -= diff
			break
		}

		widths["CALLER"] -= 10
		total -= 10
		if total < w {
			break
		}

		widths["PATH"] -= 5
		total -= 5
		if total < w {
			break
		}
	}

	var configs []table.ColumnConfig
	for k, v := range widths {
		configs = append(
			configs,
			table.ColumnConfig{
				Name:     k,
				WidthMax: v,
			},
		)
	}

	return configs
}

const defaultWidth = 100

func termWidth() int {
	w := term.Width()
	if w != 0 {
		return w
	}
	col := os.Getenv("COLUMNS")
	if col != "" {
		w, err := strconv.Atoi(col)
		if err != nil && w > 0 {
			return w
		}
	}
	return defaultWidth
}
