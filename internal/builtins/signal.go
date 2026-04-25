/*
 *
 * RR2 - internal/builtins/signal.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/context"
	"chip-go/internal/errors"
	"chip-go/internal/orchestrator"
	"chip-go/internal/values"
	"os"
	"sync"
	"syscall"
)

// Signal handling runs a forwarder goroutine that drains the raw os/signal
// channel into a managed queue, then broadcasts a wakeup to every actor blocked
// in signal(). Consumers pop from the queue atomically with their state
// transition, closing the race with CheckDeadlock.
//
// Lock order: orchestrator.mu before signalMu.
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

// ensureSignalInfra lazily wires up the signal infrastructure. Caller holds
// signalMu.
func ensureSignalInfra() {
	if signalCh != nil {
		return
	}

	signalCh = make(chan os.Signal, 16)
	caught = make(map[syscall.Signal]bool)
	signalWaiters = make(map[*orchestrator.Instance]chan struct{})

	go signalForwarder()
}

func signalForwarder() {
	for sig := range signalCh {
		signalMu.Lock()
		signalQueue = append(signalQueue, sig)

		for _, wakeup := range signalWaiters {
			select {
			case wakeup <- struct{}{}:
			default:
			}
		}

		signalMu.Unlock()
	}
}

// tryConsumeSignal atomically pops the oldest queued signal and transitions
// inst to StateRunning. Returns (zero, false) when the queue is empty, with
// inst's state untouched.
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
		got = true
		return true
	})

	return sig, got
}

func signalFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCount("signal", 0),
			ctx,
		))
	}

	signalMu.Lock()
	ensureSignalInfra()
	signalMu.Unlock()

	orch := orchestrator.Get()
	instanceID := context.GetInstanceID(ctx)
	inst := orch.GetInstance(instanceID)

	if inst == nil {
		return res.Failure(errors.NewRTError(
			nil, nil,
			shared.Errors.InvalidValue("Invalid actor handle"),
			ctx,
		))
	}

	// Fast path: signal already queued.
	if sig, ok := tryConsumeSignal(inst); ok {
		return res.Success(values.NewNumber(float64(signalToInt(sig))).SetContext(ctx))
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
		return res.Success(values.NewNumber(float64(signalToInt(sig))).SetContext(ctx))
	}

	cancelCh := orch.BeginBlocking(inst, orchestrator.StateBlockedSignal)
	orch.CheckDeadlock()

	for {
		select {
		case <-wakeup:
			if sig, ok := tryConsumeSignal(inst); ok {
				return res.Success(values.NewNumber(float64(signalToInt(sig))).SetContext(ctx))
			}

		case <-cancelCh:
			// A signal may have landed during the cancel race.
			if sig, ok := tryConsumeSignal(inst); ok {
				return res.Success(values.NewNumber(float64(signalToInt(sig))).SetContext(ctx))
			}

			orch.EndBlocking(inst)

			return res.Failure(errors.NewRTError(
				nil, nil,
				shared.Errors.InvalidValue("Deadlock: signal() blocked with no signals being caught"),
				ctx,
			))
		}
	}
}

func signalToInt(sig os.Signal) int {
	if s, ok := sig.(syscall.Signal); ok {
		return int(s)
	}

	return 0
}
