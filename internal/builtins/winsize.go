/*
 *
 * RR2 - internal/builtins/winsize.go
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

func winsizeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("winsize", 0))
	}

	fd := int(os.Stdin.Fd())

	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	rows := values.NewNumber(ws.Row).SetContext(ctx)
	cols := values.NewNumber(ws.Col).SetContext(ctx)

	result := values.NewMapFromEntries(
		[]string{"rows", "columns"},
		map[string]values.Value{
			"rows":    rows,
			"columns": cols,
		},
	)

	return res.Success(result.SetContext(ctx))
}
