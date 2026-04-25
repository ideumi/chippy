/*
 *
 * RR2 - internal/orchestrator/state.go
 *
 */

package orchestrator

type ActorState int

const (
	StateRunning ActorState = iota
	StateBlockedReceive
	StateBlockedWait
	StateBlockedSignal
	StateFinished
)
