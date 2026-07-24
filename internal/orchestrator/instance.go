/*
 *
 * RR2 - internal/orchestrator/instance.go
 *
 */

package orchestrator

import (
	"chip-go/internal/handles"
	"chip-go/internal/values"
)

type Modena interface {
	Run(filename, text string) (values.Value, error)
	GetGlobalContext() values.Ctx
}

type ModenaFactory func(instanceID int) Modena

type ActorResult struct {
	Value values.Value
	Err   error
}

// Instance is an actor's bookkeeping. Every field below ResultCh is guarded by
// Orchestrator.mu.
type Instance struct {
	ID       int
	Modena   Modena
	Registry *handles.HandleRegistry
	Inbox    *Inbox
	ResultCh chan ActorResult

	State      ActorState
	cancelCh   chan struct{}
	cancelled  bool
	waited     bool
	loadedOpts []string
}

// SendResult never blocks: ResultCh is buffered(1).
func (inst *Instance) SendResult(result ActorResult) {
	inst.ResultCh <- result
}
