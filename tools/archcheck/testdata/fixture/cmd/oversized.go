package cmd

// Hello is a stub used by the fixture.
func Hello() string { return "hello" }

// OversizedHandler intentionally exceeds the 10 effective-line cmd-func-size limit.
func OversizedHandler() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
	g := 7
	h := 8
	i := 9
	j := 10
	k := 11
	_ = a + b + c + d + e + f + g + h + i + j + k
}
