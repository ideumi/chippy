/*
 *
 * RR2 - internal/builtins/signal.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	signalHandlers = make(map[os.Signal]*values.Function)
	signalChans    = make(map[os.Signal]chan os.Signal)
	signalCancels  = make(map[os.Signal]context.CancelFunc)
	signalMutex    sync.RWMutex
)

// resolveSignalNumber maps signal numbers to os.Signal types
func resolveSignalNumber(signum int) (os.Signal, error) {
	switch signum {
	case int(syscall.SIGHUP):
		return syscall.SIGHUP, nil
	case int(syscall.SIGINT):
		return syscall.SIGINT, nil
	case int(syscall.SIGQUIT):
		return syscall.SIGQUIT, nil
	case int(syscall.SIGUSR1):
		return syscall.SIGUSR1, nil
	case int(syscall.SIGUSR2):
		return syscall.SIGUSR2, nil
	case int(syscall.SIGPIPE):
		return syscall.SIGPIPE, nil
	case int(syscall.SIGALRM):
		return syscall.SIGALRM, nil
	case int(syscall.SIGTERM):
		return syscall.SIGTERM, nil
	case int(syscall.SIGCHLD):
		return syscall.SIGCHLD, nil
	case int(syscall.SIGCONT):
		return syscall.SIGCONT, nil
	case int(syscall.SIGTSTP):
		return syscall.SIGTSTP, nil
	case int(syscall.SIGTTIN):
		return syscall.SIGTTIN, nil
	case int(syscall.SIGTTOU):
		return syscall.SIGTTOU, nil
	case int(syscall.SIGURG):
		return syscall.SIGURG, nil
	case int(syscall.SIGWINCH):
		return syscall.SIGWINCH, nil
	default:
		return nil, fmt.Errorf("Unsupported signal number")
	}
}

func signalFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 2 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("signal", 2, "signum, handlerFunc"),
			ctx,
		))
	}

	signumNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("signal", shared.PositionFirst, shared.TypeNumber, "signum"),
			ctx,
		))
	}

	handlerFunc, ok := args[1].(*values.Function)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("signal", shared.PositionSecond, shared.TypeFunction, "handlerFunc"),
			ctx,
		))
	}

	signum := int(signumNum.Value)
	sig, err := resolveSignalNumber(signum)
	if err != nil {
		posStart, posEnd := args[0].GetPos()
		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			err.Error(),
			ctx,
		))
	}

	signalMutex.Lock()
	defer signalMutex.Unlock()

	// Stop previous handler for this signal if exists
	if oldChan, exists := signalChans[sig]; exists {
		// First, cancel the goroutine to allow it to process pending signals
		if cancelFunc, exists := signalCancels[sig]; exists {
			cancelFunc()
		}

		// Give goroutine time to drain any buffered signals and exit cleanly
		// This prevents signal loss and ensures proper goroutine termination
		go func(ch chan os.Signal) {
			// Brief delay to allow goroutine cleanup
			select {
			case <-ch:
				// Drain any final signal
			default:
			}
			signal.Stop(ch)
			close(ch)
		}(oldChan)
	}

	// Store handler function
	signalHandlers[sig] = handlerFunc

	// Create new channel and register handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sig)
	signalChans[sig] = sigChan

	// Create context for this goroutine
	signalCtx, cancel := context.WithCancel(context.Background())
	signalCancels[sig] = cancel

	// Start goroutine to handle the signal
	go func() {
		defer func() {
			// Clean up when goroutine terminates
			signalMutex.Lock()
			delete(signalCancels, sig)
			signalMutex.Unlock()
		}()

		for {
			select {
			case <-signalCtx.Done():
				// Context cancelled, drain any remaining signals before exit
				for {
					select {
					case remainingSig := <-sigChan:
						if remainingSig != nil {
							signalMutex.RLock()
							handler := signalHandlers[sig]

							if handler != nil {
								result := handler.Execute([]values.Value{})
								if result.Error != nil {
									fmt.Fprintf(os.Stderr, "Signal handler error during shutdown: %s\n", result.Error.Error())
								}
							}

							signalMutex.RUnlock()
						}
					default:
						// No more signals to drain
						return
					}
				}
			case receivedSig := <-sigChan:
				if receivedSig != nil {
					signalMutex.RLock()
					handler := signalHandlers[sig]

					if handler != nil {
						result := handler.Execute([]values.Value{})
						if result.Error != nil {
							fmt.Fprintf(os.Stderr, "Signal handler error: %s\n", result.Error.Error())
						}
					}

					signalMutex.RUnlock()
				}
			}
		}
	}()

	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}

func unsignalFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("unsignal", 1, "signum"),
			ctx,
		))
	}

	signumNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("unsignal", shared.PositionFirst, shared.TypeNumber, "signum"),
			ctx,
		))
	}

	signum := int(signumNum.Value)
	sig, err := resolveSignalNumber(signum)
	if err != nil {
		posStart, posEnd := args[0].GetPos()
		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			err.Error(),
			ctx,
		))
	}

	signalMutex.Lock()
	defer signalMutex.Unlock()

	// Check if handler exists for this signal
	if oldChan, exists := signalChans[sig]; exists {
		// Cancel the goroutine
		if cancelFunc, exists := signalCancels[sig]; exists {
			cancelFunc()
		}

		// Clean up background channel closure
		go func(ch chan os.Signal, signal_to_reset os.Signal) {
			// Brief delay to allow goroutine cleanup
			select {
			case <-ch:
				// Drain any final signal
			default:
			}
			// Reset signal to OS default behavior
			signal.Reset(signal_to_reset)
			close(ch)
		}(oldChan, sig)

		// Remove from our maps
		delete(signalHandlers, sig)
		delete(signalChans, sig)
		delete(signalCancels, sig)

		return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
	}

	// Signal was not registered, return success anyway (idempotent)
	return res.Success(values.NewString(constants.STR_OK).SetContext(ctx))
}
