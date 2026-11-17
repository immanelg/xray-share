package main
// https://github.com/jhunt/go-ansi/
/* 
The MIT License (MIT)

Copyright (c) 2018 James Hunt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to
deal in the Software without restriction, including without limitation the
rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software..

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
IN THE SOFTWARE.
*/

import (
	"regexp"
	"unicode"
)

var (
	colors = map[string]string{
		"k": "00;30", // black
		"K": "01;30", // black (BOLD)

		"r": "00;31", // red
		"R": "01;31", // red (BOLD)

		"g": "00;32", // green
		"G": "01;32", // green (BOLD)

		"y": "00;33", // yellow
		"Y": "01;33", // yellow (BOLD)

		"b": "00;34", // blue
		"B": "01;34", // blue (BOLD)

		"m": "00;35", // magenta
		"M": "01;35", // magenta (BOLD)
		"p": "00;35", // magenta
		"P": "01;35", // magenta (BOLD)

		"c": "00;36", // cyan
		"C": "01;36", // cyan (BOLD)

		"w": "00;37", // white
		"W": "01;37", // white (BOLD)
	}

	re = regexp.MustCompile(`(?s)@[kKrRgGyYbBmMpPcCwW*]{.*?}`)
	st = regexp.MustCompile("\033\\[\\d+(;\\d+)?m")
)

func ansi(s string) string {
	return re.ReplaceAllStringFunc(s, func(m string) string {
		if m[1:2] == "*" {
			rainbow := "RYGCBM"
			s := ""
			i := 0
			for _, c := range m[3 : len(m)-1] {
				if unicode.IsSpace(c) {
					s += string(c)
					continue
				}

				j := i % len(rainbow)
				s += "\033[" + colors[rainbow[j:j+1]] + "m" + string(c) + "\033[00m"
				i++
			}
			return s
		}
		return "\033[" + colors[m[1:2]] + "m" + m[3:len(m)-1] + "\033[00m"
	})
}
