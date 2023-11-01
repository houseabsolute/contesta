package contesta

// type SliceTester struct {
// 	tests []SliceTest
// }

// type SliceTest struct {
// 	index int
// 	test  any
// }

// type IncompleteSliceTest struct {
// 	index  int
// 	isNext bool
// }

// func Slice(test SliceTest, tests ...SliceTest) SliceTester {
// 	return SliceTester{append([]SliceTest{test}, tests...)}
// }

// func Next() IncompleteSliceTest {
// 	return IncompleteSliceTest{isNext: true}
// }

// func (ist IncompleteSliceTest) Is(test any) SliceTest {
// 	return SliceTest{ist.index, test}
// }

// func (st SliceTest) Test(c *C, value any) Failure {
// 	return c.is(value, st.test)
// }
