/*
 *
 * RR2 - internal/builtins/winsize.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"

	"golang.org/x/sys/unix"
)

func winsizeFunction(args []values.Value, ctx values.Ctx) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("winsize", 0)))
	}

	fd := int(os.Stdin.Fd())

	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	rows := values.NewNumber(ws.Row).SetContext(ctx)
	cols := values.NewNumber(ws.Col).SetContext(ctx)

	return res.Success(values.NewList([]values.Value{rows, cols}).SetContext(ctx))
}
