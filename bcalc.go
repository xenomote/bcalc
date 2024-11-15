package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

)

type tok struct {
	s string
	t rune // o, (, ), n
}

type num struct {
	v float64
	u rune
}

func (n num) String() string {
	s := fmt.Sprint(n.v)
	x := ""
	switch n.u {		
	case 'b':
		x = "B"
	case 'k':
		x = "KB"
	case 'm':
		x = "MB"
	case 'g':
		x = "GB"
	case 't':
		x = "TB"
	case 'p':
		x = "PB"
	}
	return s+x
}

/*

U := ([kKmMgGtTpP][iI]?)?[bB]
N := [+-]?[0-9]+(\.[0-9]+([eE][+-]?[0-9]+(\.[0-9]+)?)?)?
V := N U?
C := V (in U)?
E := C ([/*+-] C)*

*/

func main() {
	ss := strings.Fields(strings.Join(os.Args[1:], " "))

	t := toks{}

	for _, s := range ss {
		t = append(t, tok{s: s, t: 0})
	}

	v := t.exp()

	pass(v.String())
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



func (t *toks) exp() num {
	if t.end() {
		fail("unexpected end of input")
	}

	v := t.val()

	n := 0
	var vs []num
	var os []string
	for !t.end() && t.head().s != ")" {
		os = append(os, t.op())
		vs = append(vs, t.val())
		n++
	}

	for i := range n {
		o := os[i]
		u := vs[i]

		switch o {
		case "+":
			if (v.u == 0) != (u.u == 0) {
				fail("cannot add values with and without units -", u)
			}

			if v.u == u.u {
				v.v += u.v
				continue
			}

			fail("add not implemented")

		case "-":
			if (v.u == 0) != (u.u == 0) {
				fail("cannot subtract values with and without units")
			}

			if v.u == u.u {
				v.v -= u.v
				continue
			}

			fail("subtract not implemented")

		case "x", "*":
			if v.u != 0 && u.u != 0 {
				fail("cannot multiply two byte units")
			}
			
			v.v *= u.v
			if v.u == 0 {
				v.u = u.u
			}

		case "/":
			if v.u == 0 && u.u != 0 {
				fail("cannot divide a unitless number")
			}
			
			v.v /= u.v
			if v.u == u.u {
				v.u = 0
			}
		}
	}

	return v
}

func (t *toks) val() num {
	x := t.head()
	if x.s != "(" {
		return t.num()
	}

	t.open()
	v := t.exp()
	t.close()

	return v
}

func (t *toks) op() string {
	if t.end() {
		fail("unexpected end of input - expected operator")
	}

	x := t.head()
	switch x.s {
	case "+":
	case "-":
	case "x", "*":
	case "/":
	default:
		fail("unexpected token", x.s, "- expected operator")
	}

	t.step()
	return x.s
}

func (t *toks) open() {
	if t.end() {
		fail("unexpected end of input - expected open bracket")
	}

	x := t.head()
	if x.s != "(" {
		fail("unexpected token", x.s, "- expected open bracket")
	}

	t.step()
}

func (t *toks) close() {
	if t.end() {
		fail("unexpected end of input - expected close bracket")
	}

	x := t.head()
	if x.s != ")" {
		fail("unexpected token", x.s, "- expected close bracket")
	}

	t.step()
}

func (t *toks) num() num {
	if t.end() {
		fail("unexpected end of input, expected number")
	}

	x := t.head()

	cs := []rune(x.s)
	i := len(cs)
	for ; i >= 0; i-- {
		if !unicode.IsLetter(cs[i - 1]) {
			break
		}
	}

	value, unit := string(cs[:i]), string(cs[i:])

	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		fail(err.Error())
	}

	u := rune(0)
	switch strings.ToLower(unit) {
	case "":
	case "b":
		u = 'b'
	case "kb":
		u = 'k'
	case "kib":
		u = 'K'
	case "mb": 
		u = 'm'
	case "mib":
		u = 'M'
	case "gb":
		u = 'g'
	case "gib":
		u = 'G'
	case "tb":
		u = 't'
	case "tib":
		u = 'T'
	case "pb":
		u = 'p'
	case "pib":
		u = 'P'
	default:
		fail("unknown unit", unit)
	}

	t.step()

	return num{v: v, u: u}
}

func pass(msg... any) {
	fmt.Println(msg...)
	os.Exit(0)
}

func fail(msg... any){
	fmt.Println(msg...)
	os.Exit(1)
}