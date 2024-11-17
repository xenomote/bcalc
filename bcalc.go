package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	t := toks{buff: []rune(strings.Join(os.Args[1:], " "))}
	t.next() // load first token

	v := t.exp()
	if !t.isEnd() {
		t.fail("malformed, couldn't parse")
	}

	fmt.Println(v.String())
}

type num struct {
	val  float64
	unit string
}

func (n num) String() string {
	return strconv.FormatFloat(n.val, 'f', -1, 64) + strings.ReplaceAll(strings.ToUpper(n.unit), "I", "i")
}

type toks struct {
	text string
	buff []rune
	n int
}

func (t *toks) isEnd() bool {
	return len(t.text) == 0
}

func (t *toks) head() string {
	return t.text
}

func (t *toks) next() {
	i := 0
	j := 0
	a := func() rune { return t.buff[t.n + i] }
	b := func() rune { return t.buff[t.n + j] }

	ok := func(x int) bool { return x + t.n < len(t.buff) }

	for ok(i) && unicode.IsSpace(a()) {
		i++
	}
	if !ok(i) {
		t.text = ""
		return
	}

	j = i + 1

	if slices.Contains([]rune("()+-*x/"), a()) {

	} else if unicode.IsDigit(a()) {
		e := false
		for ; ok(j); j++ {
			b := b()

			// e notation is allowed
			if b == 'e' || b == 'E' {
				e = true
				continue
			}

			// plus and minus must be preceeded by e 
			if (b == '+' || b == '-') && e {
				e = false
				continue
			}

			// only expect digits and decimals
			if unicode.IsDigit(b) || b == '.' {
				e = false
				continue
			}
			
			// anything else is a token end
			break
		}
	} else if unicode.IsLetter(a()) {
		for ok(j) && (unicode.IsLetter(b()) || unicode.IsNumber(b())) {
			j++
		}
	}

	t.text = string(t.buff[t.n+i:t.n+j])
	t.n += j
}

func (t *toks) dump() string {
	s := fmt.Sprint("[", t.text, "] : ")
	u := strings.Repeat(" ", len(s) + t.n - 1)
	s += string(t.buff)
	u += "^"

	return fmt.Sprint(s, "\n", u)
}

/*

E := C ([/*+-] C)*
C := V (in U)?
V := Q | \( E \)
Q := N U?
N := [+-]?[0-9]+(\.[0-9]+([eE][+-]?[0-9]+(\.[0-9]+)?)?)?
U := ([kKmMgGtTpP][iI]?)?[bB]

*/

func (t *toks) fail(s... any) {
	fmt.Println(s...)
	fmt.Println(t.dump())
	os.Exit(1)
}

func (t *toks) expect(s string) {
	if t.isEnd() {
		t.fail("unexpected end of input - expected", s)
	}
}

func (t *toks) exp() num {
	t.expect("an expression")

	v := t.conv()

	addsub := false
	mul := false
	div := false

	n := 0
	var vs []num
	var os []string
	for !t.isEnd() && t.head() != ")" {
		o := t.op()
		v := t.conv()

		if div {
			t.fail("ambiguous division - add parentheses")
		}

		switch o {
		case "/":
			if addsub || mul {
				t.fail("ambiguous division - add parentheses")
			}
			div = true

		case "*", "x":
			if addsub {
				t.fail("ambiguous multiplication - add parentheses")
			}
			mul = true

		case "+", "-":
			if mul {
				t.fail("ambiguous addition/subtraction - add parentheses")
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
				t.fail("cannot add values with and without units -", u)
			}

			v.val += u.val * m

		case "-":
			if (v.unit == "") != (u.unit == "") {
				t.fail("cannot subtract values with and without units")
			}

			v.val -= u.val * m

		case "x", "*":
			if v.unit != "" && u.unit != "" {
				t.fail("cannot multiply two byte units")
			}

			v.val *= u.val
			if v.unit == "" {
				v.unit = u.unit
			}

		case "/":
			if v.unit == "" && u.unit != "" {
				t.fail("cannot divide a unitless number")
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
		t.fail("unexpected token", x, "- expected operator")
	}

	t.next()
	return x
}

func (t *toks) conv() num {
	t.expect("an expression")

	v := t.val()

	if !t.isEnd() && t.head() == "in" {
		t.next()

		u := t.unit()
		if v.unit == "" {
			t.fail("cannot convert unitless number to", u)
		}

		v.val *= bits(v.unit) / bits(u)
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
		t.fail("unexpected token", x, "- expected a number")
	}

	t.next()
	return v
}

func (t *toks) unit() string {
	t.expect("a unit")

	x := t.head()
	if !t.isUnit() {
		t.fail("unexpected token", x, "- expected a unit")
	}

	t.next()
	return x
}

func (t *toks) isUnit() bool {
	if t.isEnd() {
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
		t.fail("unexpected token", x, "- expected open bracket")
	}

	t.next()
}

func (t *toks) close() {
	t.expect("a closing bracket")

	x := t.head()
	if x != ")" {
		t.fail("unexpected token", x, "- expected close bracket")
	}

	t.next()
}
