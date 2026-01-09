/*
 *
 * RR2 - internal/optional/tls/tlshandles.go
 *
 */

/*
 * NOTE: The mutex calls in here are intentional and are supposed
 * to mirror the implementation in chip-go/internal/builtins/shared/filehandles.go.
 * We don't do concurrency currently but this costs nothing and will
 * prevent headaches in case we ever do.
 */

package tls

import (
	"chip-go/internal/builtins/shared"
	"crypto/tls"
	"sync"
)

type TLSHandle struct {
	Conn   *tls.Conn
	Mode   string // "client" or "server"
	Closed bool
}

var (
	tlsHandles = make(map[int]*TLSHandle)
	tlsMutex   sync.RWMutex
)

func getNextTLSHandle() int {
	return shared.GetNextFileHandle()
}

func storeTLSHandle(handle int, tlsHandle *TLSHandle) {
	tlsMutex.Lock()
	tlsHandles[handle] = tlsHandle
	tlsMutex.Unlock()
}

func getTLSHandle(handle int) (*TLSHandle, bool) {
	tlsMutex.RLock()
	tlsHandle, ok := tlsHandles[handle]
	tlsMutex.RUnlock()

	return tlsHandle, ok
}

func removeTLSHandle(handle int) {
	tlsMutex.Lock()
	delete(tlsHandles, handle)
	tlsMutex.Unlock()

	shared.RecycleFileHandle(handle)
}
