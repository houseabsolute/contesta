//go:build demo

package demo

import (
	"testing"

	"github.com/houseabsolute/contesta"
)

func TestIs(t *testing.T) {
	c := contesta.New(t)

	c.Is(
		map[string]map[string]int{
			"foo": {
				"bar": 42,
			},
		},
		c.Map(
			c.Key("foo").Is(
				c.Map(
					c.Key("bar").Is(43),
				),
			),
		),
	)

	c.Is(42, 42, "42 == 42")
	c.Is(42, 43)
	c.Is(42, 42.0)
	c.Is(42, "foo")
	c.Is(42, c.Map(c.Key("foo").Is(42)))
	c.Is(
		map[string]int{"foo": 43},
		c.Map(
			c.Key("foo").Is(42),
		),
	)

	c.ValueIs(42, 42)
	c.ValueIs(42, 42.0)
	c.ValueIs(42, 43)
	c.ValueIs(42, 43.0)
	c.ValueIs(42, 43.1)
	c.ValueIs(42, c.Map(c.Key("foo").Is(42)))
}
