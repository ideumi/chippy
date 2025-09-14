/*
 *
 * RR2 - internal/builtins/shared/sockethandles.go
 *
 */

package shared

import (
	"net"
	"sync"
)

type SocketHandle struct {
	Conn     net.Conn     // For TCP connections (client and accepted connections)
	Listener net.Listener // For listening sockets
	UdpConn  *net.UDPConn // For UDP connections
	Mode     string       // "tcp", "udp", "listen"
	Address  string       // Remote address for UDP
}

var (
	socketHandles = make(map[int]*SocketHandle)
	socketMutex   sync.RWMutex
)

// Get next available handle from recycling system
func GetNextSocketHandle() int {
	return GetNextFileHandle()
}

func StoreSocketHandle(handle int, socket *SocketHandle) {
	socketMutex.Lock()
	socketHandles[handle] = socket
	socketMutex.Unlock()
}

func GetSocketHandle(handle int) (*SocketHandle, bool) {
	socketMutex.RLock()
	socket, ok := socketHandles[handle]
	socketMutex.RUnlock()

	return socket, ok
}

func RemoveSocketHandle(handle int) {
	socketMutex.Lock()
	delete(socketHandles, handle)
	socketMutex.Unlock()
}
