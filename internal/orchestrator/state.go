/*
 *
 * Chippy - internal/orchestrator/state.go
 *
 */

package orchestrator

type ActorState int

const (
	StateRunning  ActorState = 0
	StateBlocked  ActorState = 1
	StateFinished ActorState = 2
)
