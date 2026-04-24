/*
 *
 * RR2 - internal/orchestrator/deadlock.go
 *
 */

package orchestrator

// signalProbe reports whether any OS signals are being caught or queued.
// internal/builtins/signal.go installs this via SetSignalProbe to avoid an
// import cycle. When it returns true, actors blocked in signal() are treated
// as potentially runnable.
var signalProbe func() bool

func SetSignalProbe(fn func() bool) {
	signalProbe = fn
}

// BeginBlocking marks inst as entering a blocking primitive and returns a fresh
// cancel channel. Callers must select on it so the deadlock detector can wake
// them. Always pair with EndBlocking.
func (o *Orchestrator) BeginBlocking(inst *Instance, state ActorState) chan struct{} {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.State = state
	inst.cancelCh = make(chan struct{})
	inst.cancelled = false

	return inst.cancelCh
}

func (o *Orchestrator) EndBlocking(inst *Instance) {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.State = StateRunning
	inst.cancelCh = nil
}

// MarkFinished is called by an actor goroutine after it has delivered its result.
// Must run after SendResult so a wait(handle)er racing the detector still sees
// the result.
func (o *Orchestrator) MarkFinished(inst *Instance) {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.State = StateFinished
	inst.cancelCh = nil
}

// TransitionIfTrue runs fn under Orchestrator.mu. If fn returns true, inst is
// moved to StateRunning atomically with fn's work. Lets a builtin couple a
// wakeup-consumption step (which uses its own lock) with the state change,
// closing the window where CheckDeadlock could see a stale blocked state. Any
// lock taken inside fn is acquired after Orchestrator.mu.
func (o *Orchestrator) TransitionIfTrue(inst *Instance, fn func() bool) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	if fn() {
		inst.State = StateRunning
		inst.cancelCh = nil

		return true
	}

	return false
}

// CheckDeadlock cancels every blocked instance when no instance is running and
// nothing outside the program can wake one up. Cancelled actors wake from their
// blocking primitive and return a deadlock runtime error.
//
// Must be called after every state transition that could leave the program with
// no runnable instance.
func (o *Orchestrator) CheckDeadlock() {
	o.mu.Lock()
	defer o.mu.Unlock()

	signalsCaught := false

	if signalProbe != nil {
		signalsCaught = signalProbe()
	}

	var stuck []*Instance

	for _, inst := range o.instances {
		switch inst.State {
		case StateRunning:
			return
		case StateBlockedSignal:
			if signalsCaught {
				return
			}

			stuck = append(stuck, inst)
		case StateBlockedReceive:
			// Pending items mean the actor is about to wake; treat
			// as runnable so a wait(handle)er racing the drain isnt
			// falsely cancelled.
			if inst.Inbox != nil && inst.Inbox.HasItems() {
				return
			}

			stuck = append(stuck, inst)
		case StateBlockedWait:
			stuck = append(stuck, inst)
		case StateFinished:
		}
	}

	for _, inst := range stuck {
		if !inst.cancelled && inst.cancelCh != nil {
			inst.cancelled = true

			close(inst.cancelCh)
		}
	}
}
