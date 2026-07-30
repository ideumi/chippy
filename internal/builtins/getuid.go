/*
 *
 * Chippy - internal/builtins/getuid.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"os"
)

func getuidFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("getuid", 0))
	}

	uid := os.Getuid()

	return res.Success(values.NewNumber(uid))
}
