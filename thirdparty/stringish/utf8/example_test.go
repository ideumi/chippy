// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utf8_test

import (
	"fmt"

	"chip-go/thirdparty/stringish/utf8"
)

func ExampleDecodeLastRune() {
	s := "Hello, 世界"

	for len(s) > 0 {
		r, size := utf8.DecodeLastRune(s)
		fmt.Printf("%c %v\n", r, size)

		s = s[:len(s)-size]
	}
	// Output:
	// 界 3
	// 世 3
	//   1
	// , 1
	// o 1
	// l 1
	// l 1
	// e 1
	// H 1
}

func ExampleDecodeRune() {
	s := "Hello, 世界"

	for len(s) > 0 {
		r, size := utf8.DecodeRune(s)
		fmt.Printf("%c %v\n", r, size)

		s = s[size:]
	}
	// Output:
	// H 1
	// e 1
	// l 1
	// l 1
	// o 1
	// , 1
	//   1
	// 世 3
	// 界 3
}

func ExampleEncodeRune() {
	r := '世'
	buf := make([]byte, 3)

	n := utf8.EncodeRune(buf, r)

	fmt.Println(buf)
	fmt.Println(n)
	// Output:
	// [228 184 150]
	// 3
}

func ExampleEncodeRune_outOfRange() {
	runes := []rune{
		// Less than 0, out of range.
		-1,
		// Greater than 0x10FFFF, out of range.
		0x110000,
		// The Unicode replacement character.
		utf8.RuneError,
	}
	for i, c := range runes {
		buf := make([]byte, 3)
		size := utf8.EncodeRune(buf, c)
		fmt.Printf("%d: %d %[2]s %d\n", i, buf, size)
	}
	// Output:
	// 0: [239 191 189] � 3
	// 1: [239 191 189] � 3
	// 2: [239 191 189] � 3
}

func ExampleFullRune() {
	s := string([]byte{228, 184, 150}) // 世
	fmt.Println(utf8.FullRune(s))
	fmt.Println(utf8.FullRune(s[:2]))
	// Output:
	// true
	// false
}

func ExampleRuneCount() {
	s := []byte("Hello, 世界")
	fmt.Println("bytes =", len(s))
	fmt.Println("runes =", utf8.RuneCount(s))
	// Output:
	// bytes = 13
	// runes = 9
}

func ExampleRuneLen() {
	fmt.Println(utf8.RuneLen('a'))
	fmt.Println(utf8.RuneLen('界'))
	// Output:
	// 1
	// 3
}

func ExampleRuneStart() {
	s := "a界"
	fmt.Println(utf8.RuneStart(s[0]))
	fmt.Println(utf8.RuneStart(s[1]))
	fmt.Println(utf8.RuneStart(s[2]))
	// Output:
	// true
	// true
	// false
}

func ExampleValid() {
	valid := "Hello, 世界"
	invalid := string([]byte{0xff, 0xfe, 0xfd})

	fmt.Println(utf8.Valid(valid))
	fmt.Println(utf8.Valid(invalid))
	// Output:
	// true
	// false
}

func ExampleValidRune() {
	valid := 'a'
	invalid := rune(0xfffffff)

	fmt.Println(utf8.ValidRune(valid))
	fmt.Println(utf8.ValidRune(invalid))
	// Output:
	// true
	// false
}

func ExampleAppendRune() {
	buf1 := utf8.AppendRune(nil, 0x10000)
	buf2 := utf8.AppendRune([]byte("init"), 0x10000)
	fmt.Println(string(buf1))
	fmt.Println(string(buf2))
	// Output:
	// 𐀀
	// init𐀀
}
