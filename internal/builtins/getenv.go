/*
 *
 * Chippy - internal/builtins/getenv.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
)

func getenvFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("getenv", 1, "variable name"))
	}

	nameStr, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1,
			shared.Errors.InvalidArgTypeWithHint("getenv", shared.TypeString, "variable name"))
	}

	value, exists := os.LookupEnv(nameStr.Value)

	if !exists {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewString(value))
}
