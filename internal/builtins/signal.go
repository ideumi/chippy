/*
 *
 * RR2 - internal/builtins/signal.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os"
	"sync"
	"syscall"
)

// Signal handling runs a forwarder goroutine that drains the raw os/signal channel
// into a managed queue, then broadcasts a wakeup to every actor blocked in
// signal(). Consumers pop from the queue atomically with their state transition,
// closing the race with CheckDeadlock.
//
// Lock order: orchestrator.mu before signalMu.

const signalChanBuffer = 16

var (
	signalMu      sync.Mutex
	signalCh      chan os.Signal
	signalQueue   []os.Signal
	signalWaiters map[*orchestrator.Instance]chan struct{}
	caught        map[syscall.Signal]bool
)

func init() {
	orchestrator.SetSignalProbe(func() bool {
		signalMu.Lock()
		defer signalMu.Unlock()
		return len(caught) > 0 || len(signalQueue) > 0
	})
}

// Caller holds signalMu.
func ensureSignalInfra() {
	if signalCh != nil {
		return
	}

	signalCh = make(chan os.Signal, signalChanBuffer)
	caught = make(map[syscall.Signal]bool)
	signalWaiters = make(map[*orchestrator.Instance]chan struct{})

	go signalForwarder()
}

func signalForwarder() {
	for sig := range signalCh {
		signalMu.Lock()
		signalQueue = append(signalQueue, sig)

		// Snapshot under the lock, send after release. A waiter that
		// registers after the snapshot is covered by the post-register
		// queue re-check in signalFunction, so it cannot miss this signal.
		wakeups := make([]chan struct{}, 0, len(signalWaiters))

		for _, wakeup := range signalWaiters {
			wakeups = append(wakeups, wakeup)
		}

		signalMu.Unlock()

		for _, wakeup := range wakeups {
			select {
			case wakeup <- struct{}{}:
			default:
			}
		}
	}
}

// Pops the oldest queued signal atomically with the transition to StateRunning,
// so CheckDeadlock cannot observe a stale blocked state.
func tryConsumeSignal(inst *orchestrator.Instance) (os.Signal, bool) {
	var sig os.Signal
	got := false

	orchestrator.Get().TransitionIfTrue(inst, func() bool {
		signalMu.Lock()
		defer signalMu.Unlock()

		if len(signalQueue) == 0 {
			return false
		}

		sig = signalQueue[0]
		signalQueue = signalQueue[1:]

		// Drop the backing array once drained so walked-past slots can
		// be GC'd instead of leaking for the process lifetime.
		if len(signalQueue) == 0 {
			signalQueue = nil
		}

		got = true
		return true
	})

	return sig, got
}

func signalFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("signal", 0))
	}

	signalMu.Lock()
	ensureSignalInfra()
	signalMu.Unlock()

	orch := orchestrator.Get()
	instanceID := ctx.InstanceID
	inst := orch.GetInstance(instanceID)

	if inst == nil {
		return res.Fail(shared.Errors.InvalidValue("Invalid actor handle"))
	}

	// Fast path: signal already queued.
	if sig, ok := tryConsumeSignal(inst); ok {
		return res.Success(values.NewNumber(signalToInt(sig)))
	}

	wakeup := make(chan struct{}, 1)

	signalMu.Lock()
	signalWaiters[inst] = wakeup
	signalMu.Unlock()

	defer func() {
		signalMu.Lock()
		delete(signalWaiters, inst)
		signalMu.Unlock()
	}()

	// Re-check after registering: a signal may have been enqueued and broadcast
	// before signalWaiters saw us, leaving no wakeup queued for us while
	// signalProbe still reports liveness.
	if sig, ok := tryConsumeSignal(inst); ok {
		return res.Success(values.NewNumber(signalToInt(sig)))
	}

	cancelCh := orch.BeginBlocking(inst, orchestrator.StateBlockedSignal)
	orch.CheckDeadlock()

	for {
		select {
		case <-wakeup:
			if sig, ok := tryConsumeSignal(inst); ok {
				return res.Success(values.NewNumber(signalToInt(sig)))
			}

		case <-cancelCh:
			// A signal may have landed during the cancel race.
			if sig, ok := tryConsumeSignal(inst); ok {
				return res.Success(values.NewNumber(signalToInt(sig)))
			}

			orch.EndBlocking(inst)

			return res.Fail(
				shared.Errors.InvalidValue("Deadlock: signal() blocked with no signals being caught"))
		}
	}
}

func signalToInt(sig os.Signal) int {
	if typed, ok := sig.(syscall.Signal); ok {
		return int(typed)
	}

	return 0
}
