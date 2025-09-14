/*
 *
 * RR2 - internal/builtins/load.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
)

type RoadRunner2Interface interface {
	Run(filename, text string) (values.Value, error)
}

var globalRoadRunner2 RoadRunner2Interface

func SetGlobalRoadRunner2(rr RoadRunner2Interface) {
	globalRoadRunner2 = rr
}

func loadFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("load", 1, "filename"),
			ctx,
		))
	}

	filename, ok := args[0].(*values.String)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypeWithHint("load", shared.TypeString, "filename"),
			ctx,
		))
	}

	content, err := os.ReadFile(filename.Value)

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Failed to load file \""+filename.Value+"\": "+err.Error(),
			ctx,
		))
	}

	if globalRoadRunner2 == nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"RoadRunner2 not available",
			ctx,
		))
	}

	_, err = globalRoadRunner2.Run(filename.Value, string(content))

	if err != nil {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			"Failed to execute file \""+filename.Value+"\":\n"+err.Error(),
			ctx,
		))
	}

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
