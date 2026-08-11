/*
 *
 * Chippy - internal/orchestrator/deadlock.go
 *
 */

package orchestrator

// BeginBlocking marks inst as entering a blocking primitive and returns a fresh
// cancel channel. Callers must select on it so the deadlock detector can wake
// them. Always pair with EndBlocking.
//
// wakeable reports whether something outside inst can still make it runnable.
// It runs with Orchestrator.mu held, so any lock it takes is acquired after it.
func (o *Orchestrator) BeginBlocking(inst *Instance, wakeable func() bool) chan struct{} {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.State = StateBlocked
	inst.cancelCh = make(chan struct{})
	inst.cancelled = false
	inst.wakeable = wakeable

	o.checkDeadlock()

	return inst.cancelCh
}

func clearBlocking(inst *Instance) {
	inst.cancelCh = nil
	inst.wakeable = nil
}

func (o *Orchestrator) EndBlocking(inst *Instance) {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.State = StateRunning

	clearBlocking(inst)
}

// Must run after SendResult so a wait(handle)er racing the detector still sees
// the result.
func (o *Orchestrator) MarkFinished(inst *Instance) {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst.State = StateFinished
	inst.Modena = nil

	clearBlocking(inst)

	o.checkDeadlock()
}

// TransitionIfTrue runs fn under Orchestrator.mu. If fn returns true, inst is
// moved to StateRunning atomically with fn's work. Lets a builtin couple a
// wakeup-consumption step (which uses its own lock) with the state change,
// closing the window where the detector could see a stale blocked state. Any
// lock taken inside fn is acquired after Orchestrator.mu.
func (o *Orchestrator) TransitionIfTrue(inst *Instance, fn func() bool) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	if fn() {
		inst.State = StateRunning

		clearBlocking(inst)

		return true
	}

	return false
}

func (o *Orchestrator) Recheck() {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.checkDeadlock()
}

// checkDeadlock cancels every blocked instance when no instance is running and
// nothing can wake one up. Cancelled actors wake from their blocking primitive
// and return a deadlock runtime error.
//
// Caller holds Orchestrator.mu.
func (o *Orchestrator) checkDeadlock() {
	var stuck []*Instance

	for _, inst := range o.instances {
		if inst.State == StateRunning {
			return
		}

		if inst.State == StateFinished {
			continue
		}

		if inst.wakeable != nil && inst.wakeable() {
			return
		}

		stuck = append(stuck, inst)
	}

	for _, inst := range stuck {
		if !inst.cancelled && inst.cancelCh != nil {
			inst.cancelled = true

			close(inst.cancelCh)
		}
	}
}
