/*
 *
 * Chippy - internal/handles/sockethandles.go
 *
 */

package handles

import (
	"net"
	"sync"
)

type SocketHandle struct {
	Conn     net.Conn
	Listener net.Listener
	UdpConn  *net.UDPConn
	Mode     string
	Address  string
}

type SocketHandles struct {
	handles map[int]*SocketHandle
	mu      sync.RWMutex
}

func NewSocketHandles() *SocketHandles {
	return &SocketHandles{
		handles: make(map[int]*SocketHandle),
	}
}

func (sh *SocketHandles) Get(id int) (*SocketHandle, bool) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	socket, ok := sh.handles[id]

	return socket, ok
}

func (sh *SocketHandles) Store(id int, socket *SocketHandle) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.handles[id] = socket
}

func (sh *SocketHandles) Remove(id int) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	delete(sh.handles, id)
}

func (sh *SocketHandles) Extract(id int) (*SocketHandle, bool) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	socket, ok := sh.handles[id]

	if ok {
		delete(sh.handles, id)
	}

	return socket, ok
}

func (sh *SocketHandles) CloseAll() {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	for id, socket := range sh.handles {
		closeSocket(socket)
		delete(sh.handles, id)
	}
}

func closeSocket(socket *SocketHandle) {
	switch socket.Mode {
	case "tcp":
		if socket.Conn != nil {
			socket.Conn.Close()
		}
	case "udp":
		if socket.UdpConn != nil {
			socket.UdpConn.Close()
		}
	case "listen":
		if socket.Listener != nil {
			socket.Listener.Close()
		}
	}
}
