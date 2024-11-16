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
		t = append(t, tok{s: s, t: 0})
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
	v float64
	u string
}

func (n num) String() string {
	return fmt.Sprint(n.v) + n.u
}



type tok struct {
	s string
	t rune // o, (, ), n
}

type toks []tok

func (t *toks) end() bool {
	return len(*t) == 0
}

func (t *toks) head() tok {
	return (*t)[0]
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
	for !t.end() && t.head().s != ")" {
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

		switch o {
		case "+":
			if (v.u == "") != (u.u == "") {
				fail("cannot add values with and without units -", u)
			}

			if v.u == u.u {
				v.v += u.v
				continue
			}

			fail("add not implemented")

		case "-":
			if (v.u == "") != (u.u == "") {
				fail("cannot subtract values with and without units")
			}

			if v.u == u.u {
				v.v -= u.v
				continue
			}

			fail("subtract not implemented")

		case "x", "*":
			if v.u != "" && u.u != "" {
				fail("cannot multiply two byte units")
			}

			v.v *= u.v
			if v.u == "" {
				v.u = u.u
			}

		case "/":
			if v.u == "" && u.u != "" {
				fail("cannot divide a unitless number")
			}

			v.v /= u.v
			if v.u == u.u {
				v.u = ""
			}
		}
	}

	return v
}

func (t *toks) conv() num {
	t.expect("an expression")

	v := t.val()

	if !t.end() && t.head().s == "in" {
		t.step()
		u := t.unit()

		if v.u == "" {
			fail("cannot convert unitless number to", u)
		}

		v.u = u
	}

	return v
}

func (t *toks) val() num {
	t.expect("an open bracket or a number")

	x := t.head()
	if x.s != "(" {
		return t.lit()
	}

	t.open()
	v := t.exp()
	t.close()

	return v
}

func (t *toks) lit() num {
	t.expect("a number")

	n := num{v: t.num()}

	if t.isUnit() {
		n.u = t.unit()
	}

	return n
}

func (t *toks) op() string {
	t.expect("an operator")

	x := t.head()
	switch x.s {
	case "+", "-", "x", "*", "/":
	default:
		fail("unexpected token", x.s, "- expected operator")
	}

	t.step()
	return x.s
}

func (t *toks) open() {
	t.expect("an opening bracket")

	x := t.head()
	if x.s != "(" {
		fail("unexpected token", x.s, "- expected open bracket")
	}

	t.step()
}

func (t *toks) close() {
	t.expect("a closing bracket")

	x := t.head()
	if x.s != ")" {
		fail("unexpected token", x.s, "- expected close bracket")
	}

	t.step()
}

func (t *toks) num() float64 {
	t.expect("a number")

	x := t.head()

	v, err := strconv.ParseFloat(x.s, 64)
	if err != nil {
		fail("unexpected token", x.s, "- expected a number")
	}

	t.step()
	return v
}

func (t *toks) isUnit() bool {
	if t.end() {
		return false
	}

	x := t.head()
	u := strings.ToLower(x.s)
	switch u {
	case "b", "kb", "kib", "mb", "mib", "gb", "gib", "tb", "tib", "pb", "pib":
		return true
	}

	return false
}

func (t *toks) unit() string {
	t.expect("a unit")
	
	x := t.head()
	if !t.isUnit() {
		fail("unexpected token", x.s, "- expected a unit")
	}

	t.step()
	return strings.ReplaceAll(strings.ToUpper(x.s), "I", "i")
}