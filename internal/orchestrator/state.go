/*
 *
 * RR2 - internal/orchestrator/state.go
 *
 */

package orchestrator

type ActorState int

const (
	StateRunning        ActorState = 0
	StateBlockedReceive ActorState = 1
	StateBlockedWait    ActorState = 2
	StateBlockedSignal  ActorState = 3
	StateFinished       ActorState = 4
)
