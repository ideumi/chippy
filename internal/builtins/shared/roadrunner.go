/*
 *
 * RR2 - internal/builtins/shared/roadrunner.go
 *
 */

package shared

import (
	"chip-go/internal/context"
	"chip-go/internal/values"
)

type RoadRunner2Interface interface {
	Run(filename, text string) (values.Value, error)
	GetGlobalContext() *context.Context
}

var globalRoadRunner2 RoadRunner2Interface

func SetGlobalRoadRunner2(rr RoadRunner2Interface) {
	globalRoadRunner2 = rr
}

func GetGlobalRoadRunner2() RoadRunner2Interface {
	return globalRoadRunner2
}
