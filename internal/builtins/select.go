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

type fdInfo struct {
	fd     int
	handle int
}

func extractFD(handle int, forRead bool, dupFiles *[]*os.File, posStart, posEnd *errors.Position, ctx interface{}) (int, error) {
	// Try regular file handle first
	file, exists := shared.GetFileHandle(handle)

	if exists {
		fd := int(file.Fd())

		if fd >= unix.FD_SETSIZE {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("File descriptor exceeds system limit"),
				ctx,
			)
		}

		return fd, nil
	}

	// Try socket handle
	socket, socketExists := shared.GetSocketHandle(handle)

	if socketExists {
		fd, dupFile, err := extractSocketFD(socket, forRead, posStart, posEnd, ctx)

		if err != nil {
			return -1, err
		}

		if dupFile != nil {
			*dupFiles = append(*dupFiles, dupFile)
		}

		if fd >= unix.FD_SETSIZE {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("File descriptor exceeds system limit"),
				ctx,
			)
		}

		return fd, nil
	}

	// Try process handle
	proc, procExists := shared.GetProcessHandle(handle)

	if procExists {
		fd, err := extractProcessFD(proc, forRead, posStart, posEnd, ctx)

		if err != nil {
			return -1, err
		}

		if fd >= unix.FD_SETSIZE {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("File descriptor exceeds system limit"),
				ctx,
			)
		}

		return fd, nil
	}

	// No valid handle found
	return -1, errors.NewRTError(
		posStart, posEnd,
		shared.Errors.InvalidValue("Invalid file, socket, or process handle"),
		ctx,
	)
}

func extractSocketFD(socket *shared.SocketHandle, forRead bool, posStart, posEnd *errors.Position, ctx interface{}) (int, *os.File, error) {
	// TCP connection
	if socket.Conn != nil {
		tcpConn, ok := socket.Conn.(*net.TCPConn)

		if !ok {
			return -1, nil, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Socket connection is not TCP"),
				ctx,
			)
		}

		file, err := tcpConn.File()

		if err != nil {
			return -1, nil, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Cannot get file descriptor from TCP connection"),
				ctx,
			)
		}

		return int(file.Fd()), file, nil
	}

	// UDP connection
	if socket.UdpConn != nil {
		file, err := socket.UdpConn.File()

		if err != nil {
			return -1, nil, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Cannot get file descriptor from UDP connection"),
				ctx,
			)
		}

		return int(file.Fd()), file, nil
	}

	// TCP listener
	if socket.Listener != nil {
		if !forRead {
			// Cannot write to listener
			return -1, nil, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Cannot write to TCP listener socket"),
				ctx,
			)
		}

		tcpListener, ok := socket.Listener.(*net.TCPListener)

		if !ok {
			return -1, nil, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Listener is not TCP"),
				ctx,
			)
		}

		file, err := tcpListener.File()

		if err != nil {
			return -1, nil, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Cannot get file descriptor from TCP listener"),
				ctx,
			)
		}

		return int(file.Fd()), file, nil
	}

	// No valid connection
	return -1, nil, errors.NewRTError(
		posStart, posEnd,
		shared.Errors.InvalidValue("Socket handle has no valid connection"),
		ctx,
	)
}

func extractProcessFD(proc *shared.ProcessHandle, forRead bool, posStart, posEnd *errors.Position, ctx interface{}) (int, error) {
	if forRead {
		// For reading, use stdout
		if proc.Stdout == nil {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Process handle has no stdout"),
				ctx,
			)
		}

		stdoutFile, ok := proc.Stdout.(*os.File)

		if !ok {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Process handle stdout is not a file"),
				ctx,
			)
		}

		return int(stdoutFile.Fd()), nil
	} else {
		// For writing, use stdin
		if proc.Stdin == nil {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Process handle has no stdin"),
				ctx,
			)
		}

		stdinFile, ok := proc.Stdin.(*os.File)

		if !ok {
			return -1, errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue("Process handle stdin is not a file"),
				ctx,
			)
		}

		return int(stdinFile.Fd()), nil
	}
}

func processFdList(handlesList *values.List, forRead bool, argPos int, dupFiles *[]*os.File, args []values.Value, ctx interface{}) ([]fdInfo, *values.RuntimeResult) {
	res := values.NewRuntimeResult()
	var fdInfos []fdInfo

	for _, elem := range handlesList.Elements {

		handleNum, ok := elem.(*values.Number)

		if !ok {
			posStart, posEnd := args[argPos].GetPos()
			errorMsg := "All readFds must be numbers (file handles)"

			if !forRead {
				errorMsg = "All writeFds must be numbers (file handles)"
			}

			return nil, res.Failure(errors.NewRTError(
				posStart, posEnd,
				shared.Errors.InvalidValue(errorMsg),
				ctx,
			))
		}

		handle := int(handleNum.Value)
		posStart, posEnd := args[argPos].GetPos()

		fd, err := extractFD(handle, forRead, dupFiles, posStart, posEnd, ctx)

		if err != nil {
			return nil, res.Failure(err)
		}

		fdInfos = append(fdInfos, fdInfo{fd: fd, handle: handle})
	}

	return fdInfos, res.Success(nil)
}

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

	var dupFiles []*os.File // Track duplicated files to close after select

	// Ensure duplicated files are closed at function exit
	defer func() {
		for _, file := range dupFiles {
			file.Close()
		}
	}()

	// Process read FDs
	readFdInfos, result := processFdList(readFdsList, true, 0, &dupFiles, args, ctx)

	if result.Error != nil {
		return result
	}

	// Process write FDs
	writeFdInfos, result := processFdList(writeFdsList, false, 1, &dupFiles, args, ctx)

	if result.Error != nil {
		return result
	}

	// Build fd_sets using proper system calls
	var readFdSet, writeFdSet unix.FdSet
	var maxFd int

	// Initialize fd_sets to zero
	readFdSet.Zero()
	writeFdSet.Zero()

	for _, info := range readFdInfos {
		readFdSet.Set(info.fd)

		if info.fd > maxFd {
			maxFd = info.fd
		}
	}

	for _, info := range writeFdInfos {
		writeFdSet.Set(info.fd)

		if info.fd > maxFd {
			maxFd = info.fd
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

	// Call select
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

	for _, info := range readFdInfos {
		if readFdSet.IsSet(info.fd) {
			readyHandles = append(readyHandles, values.NewNumber(float64(info.handle)).SetContext(ctx))
		}
	}

	for _, info := range writeFdInfos {
		if writeFdSet.IsSet(info.fd) {
			readyHandles = append(readyHandles, values.NewNumber(float64(info.handle)).SetContext(ctx))
		}
	}

	return res.Success(values.NewList(readyHandles).SetContext(ctx))
}
