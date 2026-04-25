/*
 *
 * RR2 - internal/optional/tls/tlshandles.go
 *
 */

package tls

// TLS handles live in the per-actor registry at internal/handles; this file
// just provides typed accessors over the shared Optional manager.

import (
	"chip-go/internal/handles"
	"chip-go/internal/orchestrator"
	"crypto/tls"
)

type TLSHandle struct {
	Conn   *tls.Conn
	Mode   string // "client" or "server"
	Closed bool
}

func (h *TLSHandle) Close() error {
	if !h.Closed && h.Conn != nil {
		h.Closed = true
		return h.Conn.Close()
	}

	return nil
}

func getRegistry(ctx interface{}) *handles.HandleRegistry {
	return orchestrator.Get().GetRegistry(ctx)
}

func getNextTLSHandle(ctx interface{}) int {
	return getRegistry(ctx).Alloc.Alloc()
}

func storeTLSHandle(ctx interface{}, handle int, tlsHandle *TLSHandle) {
	getRegistry(ctx).Optional.Store(handle, tlsHandle)
}

func getTLSHandle(ctx interface{}, handle int) (*TLSHandle, bool) {
	opt, ok := getRegistry(ctx).Optional.Get(handle)

	if !ok {
		return nil, false
	}

	tlsHandle, ok := opt.(*TLSHandle)

	return tlsHandle, ok
}

func removeTLSHandle(ctx interface{}, handle int) {
	reg := getRegistry(ctx)
	reg.Optional.Remove(handle)
	reg.Alloc.Free(handle)
}
