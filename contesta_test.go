package contesta

import (
	"testing"
)

func TestIs(t *testing.T) {
	t.Run("simple equality", func(t *testing.T) {
		t.Run("42 == 42", func(t *testing.T) {
			m := newMockT()
			c := NewWithOutput(m, m)
			c.Is(42, 42)
			m.AssertCalled(t, "Helper")
			m.AssertPassed(t)
		})
		t.Run("42 == 43", func(t *testing.T) {
			m := newMockT()
			c := NewWithOutput(m, m)
			c.Is(42, 43)
			m.AssertCalled(t, "Helper")
			m.AssertFailed(t)
		})
	})
}
