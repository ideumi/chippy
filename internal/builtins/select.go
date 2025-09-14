/*
 *
 * RR2 - internal/builtins/select.go
 *
 */

package builtins

import (
	"chip-go/internal/builtins/shared"
	"chip-go/internal/constants"
	"chip-go/internal/errors"
	"chip-go/internal/values"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"time"
)

func selectFunction(args []values.Value, ctx interface{}) *values.RuntimeResult {
	res := values.NewRuntimeResult()

	if len(args) != 3 {
		var posStart, posEnd *errors.Position

		if len(args) > 0 {
			posStart, posEnd = args[0].GetPos()
		}

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgCountWithHint("select", 3, "readFds, writeFds, timeoutMs"),
			ctx,
		))
	}

	readFdsList, ok := args[0].(*values.List)

	if !ok {
		posStart, posEnd := args[0].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("select", shared.PositionFirst, shared.TypeList, "readFds"),
			ctx,
		))
	}

	writeFdsList, ok := args[1].(*values.List)

	if !ok {
		posStart, posEnd := args[1].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("select", shared.PositionSecond, shared.TypeList, "writeFds"),
			ctx,
		))
	}

	timeoutNum, ok := args[2].(*values.Number)

	if !ok {
		posStart, posEnd := args[2].GetPos()

		return res.Failure(errors.NewRTError(
			posStart, posEnd,
			shared.Errors.InvalidArgTypePositionalWithHint("select", shared.PositionThird, shared.TypeNumber, "timeoutMs"),
			ctx,
		))
	}

	timeout := time.Duration(timeoutNum.Value) * time.Millisecond

	var readFds []int
	var readHandles []int
	var dupFiles []*os.File // Track duplicated files to close after select

	// Ensure duplicated files are closed at function exit
	defer func() {
		for _, file := range dupFiles {
			file.Close()
		}
	}()

	for _, elem := range readFdsList.Elements {
		handleNum, ok := elem.(*values.Number)

		if !ok {
			posStart, posEnd := args[0].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("All readFds must be numbers (file handles)"),
				ctx,
			))
		}

		handle := int(handleNum.Value)

		// Try regular file handle first
		file, exists := shared.GetFileHandle(handle)

		if exists {
			fd := int(file.Fd())
			if fd >= unix.FD_SETSIZE {
				posStart, posEnd := args[0].GetPos()

				return res.Failure(errors.NewRTError(
					posStart, posEnd,
					shared.Errors.InvalidValue("File descriptor exceeds system limit"),
					ctx,
				))
			}

			readFds = append(readFds, fd)
			readHandles = append(readHandles, handle)
		} else {
			// Try socket handle
			socket, socketExists := shared.GetSocketHandle(handle)
			if socketExists {
				var fd int
				if socket.Conn != nil {
					// TCP connection
					if tcpConn, ok := socket.Conn.(*net.TCPConn); ok {
						if file, err := tcpConn.File(); err == nil {
							dupFiles = append(dupFiles, file) // Track for later cleanup
							fd = int(file.Fd())
						} else {
							posStart, posEnd := args[0].GetPos()

							return res.Failure(errors.NewRTError(
								posStart, posEnd,
								shared.Errors.InvalidValue("Cannot get file descriptor from TCP connection"),
								ctx,
							))
						}
					} else {
						posStart, posEnd := args[0].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Socket connection is not TCP"),
							ctx,
						))
					}
				} else if socket.UdpConn != nil {
					// UDP connection
					if file, err := socket.UdpConn.File(); err == nil {
						dupFiles = append(dupFiles, file) // Track for later cleanup
						fd = int(file.Fd())
					} else {
						posStart, posEnd := args[0].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Cannot get file descriptor from UDP connection"),
							ctx,
						))
					}
				} else if socket.Listener != nil {
					// TCP listener
					if tcpListener, ok := socket.Listener.(*net.TCPListener); ok {
						if file, err := tcpListener.File(); err == nil {
							dupFiles = append(dupFiles, file) // Track for later cleanup
							fd = int(file.Fd())
						} else {
							posStart, posEnd := args[0].GetPos()

							return res.Failure(errors.NewRTError(
								posStart, posEnd,
								shared.Errors.InvalidValue("Cannot get file descriptor from TCP listener"),
								ctx,
							))
						}
					} else {
						posStart, posEnd := args[0].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Listener is not TCP"),
							ctx,
						))
					}
				} else {
					posStart, posEnd := args[0].GetPos()

					return res.Failure(errors.NewRTError(
						posStart, posEnd,
						shared.Errors.InvalidValue("Socket handle has no valid connection"),
						ctx,
					))
				}

				if fd >= unix.FD_SETSIZE {
					posStart, posEnd := args[0].GetPos()

					return res.Failure(errors.NewRTError(
						posStart, posEnd,
						shared.Errors.InvalidValue("File descriptor exceeds system limit"),
						ctx,
					))
				}

				readFds = append(readFds, fd)
				readHandles = append(readHandles, handle)
			} else {
				// Try process handle
				proc, procExists := shared.GetProcessHandle(handle)

				if procExists && proc.Stdout != nil {
					// For process handles, we want to read from stdout
					if stdoutFile, ok := proc.Stdout.(*os.File); ok {
						fd := int(stdoutFile.Fd())

						if fd >= unix.FD_SETSIZE {
							posStart, posEnd := args[0].GetPos()

							return res.Failure(errors.NewRTError(
								posStart, posEnd,
								shared.Errors.InvalidValue("File descriptor exceeds system limit"),
								ctx,
							))
						}

						readFds = append(readFds, fd)
						readHandles = append(readHandles, handle)
					} else {
						posStart, posEnd := args[0].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Process handle stdout is not a file"),
							ctx,
						))
					}
				} else {
					posStart, posEnd := args[0].GetPos()

					return res.Failure(errors.NewRTError(
						posStart, posEnd,
						shared.Errors.InvalidValue("Invalid file, socket, or process handle in readFds"),
						ctx,
					))
				}
			}
		}
	}

	var writeFds []int
	var writeHandles []int

	for _, elem := range writeFdsList.Elements {
		handleNum, ok := elem.(*values.Number)
		if !ok {
			posStart, posEnd := args[1].GetPos()

			return res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("All writeFds must be numbers (file handles)"),
				ctx,
			))
		}

		handle := int(handleNum.Value)

		// Try regular file handle first
		file, exists := shared.GetFileHandle(handle)

		if exists {
			fd := int(file.Fd())
			if fd >= unix.FD_SETSIZE {
				posStart, posEnd := args[1].GetPos()

				return res.Failure(errors.NewRTError(
					posStart, posEnd,
					shared.Errors.InvalidValue("File descriptor exceeds system limit"),
					ctx,
				))
			}

			writeFds = append(writeFds, fd)
			writeHandles = append(writeHandles, handle)
		} else {
			// Try socket handle
			socket, socketExists := shared.GetSocketHandle(handle)

			if socketExists {
				var fd int
				if socket.Conn != nil {
					// TCP connection
					if tcpConn, ok := socket.Conn.(*net.TCPConn); ok {
						if file, err := tcpConn.File(); err == nil {
							dupFiles = append(dupFiles, file) // Track for later cleanup
							fd = int(file.Fd())
						} else {
							posStart, posEnd := args[1].GetPos()

							return res.Failure(errors.NewRTError(
								posStart, posEnd,
								shared.Errors.InvalidValue("Cannot get file descriptor from TCP connection"),
								ctx,
							))
						}
					} else {
						posStart, posEnd := args[1].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Socket connection is not TCP"),
							ctx,
						))
					}
				} else if socket.UdpConn != nil {
					// UDP connection
					if file, err := socket.UdpConn.File(); err == nil {
						dupFiles = append(dupFiles, file) // Track for later cleanup
						fd = int(file.Fd())
					} else {
						posStart, posEnd := args[1].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Cannot get file descriptor from UDP connection"),
							ctx,
						))
					}
				} else {
					// TCP listeners cannot be written to
					posStart, posEnd := args[1].GetPos()

					return res.Failure(errors.NewRTError(
						posStart, posEnd,
						shared.Errors.InvalidValue("Cannot write to TCP listener socket"),
						ctx,
					))
				}

				if fd >= unix.FD_SETSIZE {
					posStart, posEnd := args[1].GetPos()

					return res.Failure(errors.NewRTError(
						posStart, posEnd,
						shared.Errors.InvalidValue("File descriptor exceeds system limit"),
						ctx,
					))
				}

				writeFds = append(writeFds, fd)
				writeHandles = append(writeHandles, handle)
			} else {
				// Try process handle
				proc, procExists := shared.GetProcessHandle(handle)

				if procExists && proc.Stdin != nil {
					// For process handles, we want to write to stdin
					if stdinFile, ok := proc.Stdin.(*os.File); ok {
						fd := int(stdinFile.Fd())

						if fd >= unix.FD_SETSIZE {
							posStart, posEnd := args[1].GetPos()
							return res.Failure(errors.NewRTError(
								posStart, posEnd,
								shared.Errors.InvalidValue("File descriptor exceeds system limit"),
								ctx,
							))
						}

						writeFds = append(writeFds, fd)
						writeHandles = append(writeHandles, handle)
					} else {
						posStart, posEnd := args[1].GetPos()

						return res.Failure(errors.NewRTError(
							posStart, posEnd,
							shared.Errors.InvalidValue("Process handle stdin is not a file"),
							ctx,
						))
					}
				} else {
					posStart, posEnd := args[1].GetPos()

					return res.Failure(errors.NewRTError(
						posStart, posEnd,
						shared.Errors.InvalidValue("Invalid file, socket, or process handle in writeFds"),
						ctx,
					))
				}
			}
		}
	}

	// Build fd_sets using proper system calls
	var readFdSet, writeFdSet unix.FdSet
	var maxFd int

	// Initialize fd_sets to zero
	readFdSet.Zero()
	writeFdSet.Zero()

	for _, fd := range readFds {
		readFdSet.Set(fd)

		if fd > maxFd {
			maxFd = fd
		}
	}

	for _, fd := range writeFds {
		writeFdSet.Set(fd)

		if fd > maxFd {
			maxFd = fd
		}
	}

	// Prepare timeout for syscall
	var tv *unix.Timeval

	if timeout > 0 {
		tv = &unix.Timeval{
			Sec:  int64(timeout / time.Second),
			Usec: int64((timeout % time.Second) / time.Microsecond),
		}
	}

	// Call select syscall using unix package
	n, err := unix.Select(maxFd+1, &readFdSet, &writeFdSet, nil, tv)

	if err != nil {
		return res.Success(values.NewString(constants.STR_ERR).SetContext(ctx))
	}

	if n == 0 {
		// Timeout
		return res.Success(values.NewList([]values.Value{}).SetContext(ctx))
	}

	// Check which handles are ready
	var readyHandles []values.Value

	for i, fd := range readFds {
		if readFdSet.IsSet(fd) {
			readyHandles = append(readyHandles, values.NewNumber(float64(readHandles[i])))
		}
	}

	for i, fd := range writeFds {
		if writeFdSet.IsSet(fd) {
			readyHandles = append(readyHandles, values.NewNumber(float64(writeHandles[i])))
		}
	}

	return res.Success(values.NewList(readyHandles).SetContext(ctx))
}
