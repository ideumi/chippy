/*
 *
 * Chippy - internal/builtins/getterm.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"

	"golang.org/x/sys/unix"
)

// Scalar (non control character) keys surfaced by the getterm / setterm map.
// See: /usr/include/x86_64-linux-gnu/bits/termios-struct.h.
var termiosScalarKeys = []string{
	"iflag",
	"oflag",
	"cflag",
	"lflag",
	"line",
	"ispeed",
	"ospeed",
}

// Linux Cc array indices for the control characters (cc) surfaced by the getterm /
// setterm map. See: /usr/include/x86_64-linux-gnu/bits/termios-c_cc.h.
var termiosCcIndices = []struct {
	name  string
	index int
}{
	{"VINTR", 0},
	{"VQUIT", 1},
	{"VERASE", 2},
	{"VKILL", 3},
	{"VEOF", 4},
	{"VTIME", 5},
	{"VMIN", 6},
	{"VSWTC", 7},
	{"VSTART", 8},
	{"VSTOP", 9},
	{"VSUSP", 10},
	{"VEOL", 11},
	{"VREPRINT", 12},
	{"VDISCARD", 13},
	{"VWERASE", 14},
	{"VLNEXT", 15},
	{"VEOL2", 16},
}

func gettermFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("getterm", 0))
	}

	fd := int(os.Stdin.Fd())

	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	keys := append([]string{}, termiosScalarKeys...)

	entries := map[string]values.Value{
		"iflag":  values.NewNumber(termios.Iflag),
		"oflag":  values.NewNumber(termios.Oflag),
		"cflag":  values.NewNumber(termios.Cflag),
		"lflag":  values.NewNumber(termios.Lflag),
		"line":   values.NewNumber(termios.Line),
		"ispeed": values.NewNumber(termios.Ispeed),
		"ospeed": values.NewNumber(termios.Ospeed),
	}

	for _, cc := range termiosCcIndices {
		keys = append(keys, cc.name)
		entries[cc.name] = values.NewNumber(termios.Cc[cc.index])
	}

	result := values.NewMapFromEntries(keys, entries)

	return res.Success(result)
}
