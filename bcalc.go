package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	ss := strings.Fields(strings.Join(os.Args[1:], " "))
	t := toks{}

	for _, s := range ss {
		t = append(t, tok{s: s})
	}

	v := t.exp()

	pass(v.String())
}

func pass(msg ...any) {
	fmt.Println(msg...)
	os.Exit(0)
}

func fail(msg ...any) {
	fmt.Println(msg...)
	os.Exit(1)
}

type num struct {
	val  float64
	unit string
}

func (n num) String() string {
	return strconv.FormatFloat(n.val, 'f', -1, 64) + strings.ReplaceAll(strings.ToUpper(n.unit), "I", "i")
}

type tok struct {
	s string
}

type toks []tok

func (t *toks) end() bool {
	return len(*t) == 0
}

func (t *toks) head() string {
	return (*t)[0].s
}

func (t *toks) step() {
	*t = (*t)[1:]
}

func (t *toks) expect(s string) {
	if t.end() {
		fail("unexpected end of input - expected", s)
	}
}

/*

E := C ([/*+-] C)*
C := V (in U)?
V := Q | \( E \)
Q := N U?
N := [+-]?[0-9]+(\.[0-9]+([eE][+-]?[0-9]+(\.[0-9]+)?)?)?
U := ([kKmMgGtTpP][iI]?)?[bB]

*/

func (t *toks) exp() num {
	t.expect("an expression")

	v := t.conv()

	addsub := false
	mul := false
	div := false

	n := 0
	var vs []num
	var os []string
	for !t.end() && t.head() != ")" {
		o := t.op()
		v := t.conv()

		if div {
			fail("ambiguous division - add parentheses")
		}

		switch o {
		case "/":
			if addsub || mul {
				fail("ambiguous division - add parentheses")
			}
			div = true

		case "*", "x":
			if addsub {
				fail("ambiguous multiplication - add parentheses")
			}
			mul = true

		case "+", "-":
			if mul {
				fail("ambiguous addition/subtraction - add parentheses")
			}
			addsub = true
		}

		os = append(os, o)
		vs = append(vs, v)
		n++
	}

	for i := range n {
		o := os[i]
		u := vs[i]

		m := bits(u.unit) / bits(v.unit)

		switch o {
		case "+":
			if (v.unit == "") != (u.unit == "") {
				fail("cannot add values with and without units -", u)
			}

			v.val += u.val * m

		case "-":
			if (v.unit == "") != (u.unit == "") {
				fail("cannot subtract values with and without units")
			}

			v.val -= u.val * m

		case "x", "*":
			if v.unit != "" && u.unit != "" {
				fail("cannot multiply two byte units")
			}

			v.val *= u.val
			if v.unit == "" {
				v.unit = u.unit
			}

		case "/":
			if v.unit == "" && u.unit != "" {
				fail("cannot divide a unitless number")
			}

			v.val /= u.val

			if u.unit != "" {
				v.unit = ""
				v.val /= m
			}
		}
	}

	return v
}

func (t *toks) op() string {
	t.expect("an operator")

	x := t.head()
	switch x {
	case "+", "-", "x", "*", "/":
	default:
		fail("unexpected token", x, "- expected operator")
	}

	t.step()
	return x
}

func (t *toks) conv() num {
	t.expect("an expression")

	v := t.val()

	if !t.end() && t.head() == "in" {
		t.step()

		u := t.unit()
		if v.unit == "" {
			fail("cannot convert unitless number to", u)
		}

		v.unit = u
	}

	return v
}

func (t *toks) val() num {
	t.expect("an open bracket or a number")

	x := t.head()
	if x != "(" {
		return t.lit()
	}

	t.open()
	v := t.exp()
	t.close()

	return v
}

func (t *toks) lit() num {
	t.expect("a number")

	n := num{val: t.num()}

	if t.isUnit() {
		n.unit = t.unit()
	}

	return n
}

func (t *toks) num() float64 {
	t.expect("a number")

	x := t.head()

	v, err := strconv.ParseFloat(x, 64)
	if err != nil {
		fail("unexpected token", x, "- expected a number")
	}

	t.step()
	return v
}

func (t *toks) unit() string {
	t.expect("a unit")

	x := t.head()
	if !t.isUnit() {
		fail("unexpected token", x, "- expected a unit")
	}

	t.step()
	return x
}

func (t *toks) isUnit() bool {
	if t.end() {
		return false
	}

	x := t.head()
	u := strings.ToLower(x)
	switch u {
	case "b", "kb", "kib", "mb", "mib", "gb", "gib", "tb", "tib", "pb", "pib":
		return true
	}

	return false
}

func bits(s string) float64 {
	switch s {
	case "b":
		return 1
	case "kb":
		return 1_000
	case "kib":
		return 1024
	case "mb":
		return 1_000_000
	case "mib":
		return 1024 * 1024
	case "gb":
		return 1_000_000_000
	case "gib":
		return 1024 * 1024 * 1024
	case "tb":
		return 1_000_000_000_000
	case "tib":
		return 1024 * 1024 * 1024 * 1024
	case "pb":
		return 1_000_000_000_000_000
	case "pib":
		return 1024 * 1024 * 1024 * 1024 * 1024
	}
	return 1
}

func (t *toks) open() {
	t.expect("an opening bracket")

	x := t.head()
	if x != "(" {
		fail("unexpected token", x, "- expected open bracket")
	}

	t.step()
}

func (t *toks) close() {
	t.expect("a closing bracket")

	x := t.head()
	if x != ")" {
		fail("unexpected token", x, "- expected close bracket")
	}

	t.step()
}
