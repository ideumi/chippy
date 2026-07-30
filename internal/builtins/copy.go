/*
 *
 * RR2 - internal/builtins/copy.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
)

func copyFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("copy", 1, "value"))
	}

	return res.Success(args[0].Copy())
}
