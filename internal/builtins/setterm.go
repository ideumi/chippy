/*
 *
 * Chippy - internal/builtins/setterm.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"fmt"
	"math"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

func settermFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("setterm", 1, "state"))
	}

	argMap, ok := values.AsMap(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("setterm", shared.TypeMap, "state"))
	}

	var missing []string

	for _, key := range termiosScalarKeys {
		if _, exists := argMap.Entries[key]; !exists {
			missing = append(missing, key)
		}
	}

	for _, cc := range termiosCcIndices {
		if _, exists := argMap.Entries[cc.name]; !exists {
			missing = append(missing, cc.name)
		}
	}

	if len(missing) > 0 {
		var detail string

		if len(missing) == 1 {
			detail = fmt.Sprintf("state is missing key '%s'", missing[0])
		} else {
			quoted := make([]string, len(missing))

			for i, key := range missing {
				quoted[i] = "'" + key + "'"
			}

			detail = fmt.Sprintf("state is missing keys: %s", strings.Join(quoted, ", "))
		}

		return res.Fail(shared.Errors.InvalidValue(detail))
	}

	iflag, rtErr := requireTermiosField(argMap, "iflag", math.MaxUint32)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	oflag, rtErr := requireTermiosField(argMap, "oflag", math.MaxUint32)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	cflag, rtErr := requireTermiosField(argMap, "cflag", math.MaxUint32)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	lflag, rtErr := requireTermiosField(argMap, "lflag", math.MaxUint32)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	line, rtErr := requireTermiosField(argMap, "line", math.MaxUint8)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	ispeed, rtErr := requireTermiosField(argMap, "ispeed", math.MaxUint32)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	ospeed, rtErr := requireTermiosField(argMap, "ospeed", math.MaxUint32)
	if rtErr != nil {
		return res.Failure(rtErr)
	}

	var termios unix.Termios

	termios.Iflag = uint32(iflag)
	termios.Oflag = uint32(oflag)
	termios.Cflag = uint32(cflag)
	termios.Lflag = uint32(lflag)
	termios.Line = uint8(line)
	termios.Ispeed = uint32(ispeed)
	termios.Ospeed = uint32(ospeed)

	for _, cc := range termiosCcIndices {
		value, rtErr := requireTermiosField(argMap, cc.name, math.MaxUint8)
		if rtErr != nil {
			return res.Failure(rtErr)
		}

		termios.Cc[cc.index] = uint8(value)
	}

	fd := int(os.Stdin.Fd())

	if err := unix.IoctlSetTermios(fd, unix.TCSETSW, &termios); err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(constants.STR_OK))
}

func requireTermiosField(termiosMap *values.Map, key string, max int64) (int64, *errors.RTError) {
	num := termiosMap.Entries[key]

	if !num.IsNumber() {
		return 0, errors.NewCallError(
			shared.Errors.InvalidValue(fmt.Sprintf("state key '%s' must be a number", key)))
	}

	intVal, err := num.AsInt()

	if err != nil {
		return 0, err.(*errors.RTError)
	}

	if intVal < 0 || intVal > max {
		return 0, errors.NewCallError(
			shared.Errors.InvalidValue(fmt.Sprintf("state key '%s' value %d is out of range [0, %d]", key, intVal, max)))
	}

	return intVal, nil
}
