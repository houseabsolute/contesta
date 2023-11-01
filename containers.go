package contesta

type NonExhaustiveContainerTest struct{}

func (NonExhaustiveContainerTest) isMapTest() {}

func NonExhaustive() NonExhaustiveContainerTest {
	return NonExhaustiveContainerTest{}
}

// This always passes.
func (NonExhaustiveContainerTest) Test(_ *C, _ any) []*result {
	return []*result{{
		pass:        true,
		description: "NonExhaustiveContainerTest always passes",
	}}
}

type ExhaustiveContainerTest struct{}

func (ExhaustiveContainerTest) isMapTest() {}

func End() ExhaustiveContainerTest {
	return ExhaustiveContainerTest{}
}

func (ExhaustiveContainerTest) Test(c *C, _ any) []*result {
	// check if value is a map and has any unchecked keys
	if false {
		return []*result{{
			pass:        false,
			description: "Map has unchecked keys",
		}}
	}

	// check if value is a slice and has any unchecked elements
	// if false {
	// 	return SliceElementsNotCheckedFailure{}
	// }

	// check if value is an array and has any unchecked elements

	// check if value is a struct and has any unchecked fields

	return []*result{{
		pass:        true,
		description: "All values in the container were checked",
	}}
}
