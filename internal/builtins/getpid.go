/*
 *
 * Chippy - internal/builtins/getpid.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"os"
)

func getpidFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("getpid", 0))
	}

	pid := os.Getpid()

	return res.Success(values.NewNumber(pid))
}
