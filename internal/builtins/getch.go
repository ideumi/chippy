/*
 *
 * Chippy - internal/builtins/getch.go
 *
 */

/*
 * https://pkg.go.dev/golang.org/x/term
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/values"
	"os"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func getchSequence() ([]byte, error) {
	fd := int(os.Stdin.Fd())

	// Save original terminal state
	oldState, err := term.GetState(fd)

	if err != nil {
		return nil, err
	}
	defer term.Restore(fd, oldState)

	// Set terminal to raw mode
	_, err = term.MakeRaw(fd)

	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 8)

	// Read first byte
	readCount, err := os.Stdin.Read(buffer[:1])

	if err != nil || readCount == 0 {
		return nil, err
	}

	bytesRead := 1
	firstByte := buffer[0]

	// Escape sequences
	if firstByte == 27 && bytesRead < len(buffer) {
		err := unix.SetNonblock(fd, true)

		if err == nil {
			defer unix.SetNonblock(fd, false)

			// More bytes
			timeout := time.NewTimer(100 * time.Millisecond)

			defer timeout.Stop()

			done := make(chan bool, 1)
			go func() {
				for bytesRead < len(buffer) {
					readCount, err := os.Stdin.Read(buffer[bytesRead : bytesRead+1])

					if err != nil || readCount == 0 {
						break
					}
					bytesRead++
				}
				done <- true
			}()

			select {
			case <-done:
				// Got additional bytes
			case <-timeout.C:
				// Timeout reached
			}
		}
	} else if (firstByte&0x80) != 0 && bytesRead < len(buffer) {
		// Handle UTF8 multi byte characters
		var additionalBytes int

		if (firstByte & 0xE0) == 0xC0 {
			additionalBytes = 1 // 2-byte UTF-8
		} else if (firstByte & 0xF0) == 0xE0 {
			additionalBytes = 2 // 3-byte UTF-8
		} else if (firstByte & 0xF8) == 0xF0 {
			additionalBytes = 3 // 4-byte UTF-8
		}

		// Ensure we don't exceed buffer
		if bytesRead+additionalBytes > len(buffer) {
			additionalBytes = len(buffer) - bytesRead
		}

		// Read additional UTF8 bytes
		for i := 0; i < additionalBytes; i++ {
			readCount, err := os.Stdin.Read(buffer[bytesRead : bytesRead+1])

			if err != nil || readCount == 0 {
				break
			}

			// Validate UTF-8 continuation byte: 10xxxxxx
			if (buffer[bytesRead] & 0xC0) != 0x80 {
				// Invalid UTF8
				break
			}
			bytesRead++
		}
	}

	return buffer[:bytesRead], nil
}

func getchFunction(args []values.Value, ctx values.Ctx) values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 0 {
		return res.Fail(shared.Errors.InvalidArgCount("getch", 0))
	}

	bytes, err := getchSequence()

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR))
	}

	return res.Success(values.NewBytes(bytes))
}
