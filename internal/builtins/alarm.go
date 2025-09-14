/*
 *
 * RR2 - internal/builtins/alarm.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"os"
	"sync"
	"syscall"
	"time"
)

var (
	alarmTimer     *time.Timer
	alarmMutex     sync.Mutex
	alarmRemaining time.Duration
	alarmStartTime time.Time
)

func alarmFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 1 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("alarm", 1, "milliseconds"),
			ctx,
		))
	}

	millisecondsNum, ok := args[0].(*values.Number)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("alarm", shared.PositionFirst, shared.TypeNumber, "milliseconds"),
			ctx,
		))
	}

	milliseconds := millisecondsNum.Value

	alarmMutex.Lock()
	defer alarmMutex.Unlock()

	var remainingMilliseconds float64 = 0

	// Calculate remaining time from previous alarm
	if alarmTimer != nil {
		elapsed := time.Since(alarmStartTime)

		if elapsed < alarmRemaining {
			remainingMilliseconds = float64((alarmRemaining - elapsed).Nanoseconds()) / 1000000
		}

		alarmTimer.Stop()
		alarmTimer = nil
	}

	// Set new alarm if milliseconds > 0
	if milliseconds > 0 {
		alarmRemaining = time.Duration(milliseconds) * time.Millisecond
		alarmStartTime = time.Now()
		alarmTimer = time.NewTimer(alarmRemaining)

		go func() {
			<-alarmTimer.C
			// Send SIGALRM to current process
			syscall.Kill(os.Getpid(), syscall.SIGALRM)

			alarmMutex.Lock()
			alarmTimer = nil
			alarmMutex.Unlock()
		}()
	}

	return res.Success(values.NewNumber(remainingMilliseconds).SetContext(ctx))
}
