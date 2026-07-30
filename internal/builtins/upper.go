/*
 *
 * RR2 - internal/builtins/upper.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/values"
	"strings"
)

func upperFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		return res.Fail(shared.Errors.InvalidArgCountWithHint("upper", 1, "string"))
	}

	stringArg, ok := values.AsString(args[0])

	if !ok {
		return res.FailAt(1, shared.Errors.InvalidArgTypeWithHint("upper", shared.TypeString, "string"))
	}

	str := stringArg.Value
	result := strings.ToUpper(str)

	return res.Success(values.NewString(result))
}
