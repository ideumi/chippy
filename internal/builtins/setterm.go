/*
 *
 * RR2 - internal/builtins/setterm.go
 *
 */

package builtins

import (
	"bytes"
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"encoding/binary"
	"os"

	"golang.org/x/sys/unix"
)

func settermFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("setterm", 1, "state"),
			ctx,
		))
	}

	argBytes, ok := args[0].(*values.Bytes)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("setterm", shared.TypeBytes, "state"),
			ctx,
		))
	}

	if len(argBytes.Data) != binary.Size(unix.Termios{}) {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	// Deserialize
	var termios unix.Termios
	reader := bytes.NewReader(argBytes.Data)

	if err := binary.Read(reader, binary.LittleEndian, &termios); err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	fd := int(os.Stdin.Fd())

	if err := unix.IoctlSetTermios(fd, unix.TCSETSW, &termios); err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
