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
	"chip-go/internal/values"
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

func getRegistry(ctx values.Ctx) *handles.HandleRegistry {
	return orchestrator.Get().GetRegistry(ctx.InstanceID)
}

func getNextTLSHandle(ctx values.Ctx) int {
	return getRegistry(ctx).Alloc.Alloc()
}

func storeTLSHandle(ctx values.Ctx, handle int, tlsHandle *TLSHandle) {
	getRegistry(ctx).Optional.Store(handle, tlsHandle)
}

func getTLSHandle(ctx values.Ctx, handle int) (*TLSHandle, bool) {
	opt, ok := getRegistry(ctx).Optional.Get(handle)

	if !ok {
		return nil, false
	}

	tlsHandle, ok := opt.(*TLSHandle)

	return tlsHandle, ok
}

func removeTLSHandle(ctx values.Ctx, handle int) {
	reg := getRegistry(ctx)
	reg.Optional.Remove(handle)
	reg.Alloc.Free(handle)
}
