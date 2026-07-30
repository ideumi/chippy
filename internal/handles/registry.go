/*
 *
 * Chippy - internal/handles/registry.go
 *
 */

package handles

type HandleManager interface {
	CloseAll()
}

type HandleRegistry struct {
	Alloc     *HandleAllocator
	Files     *FileHandles
	Dirs      *DirHandles
	Sockets   *SocketHandles
	Processes *ProcessHandles
	Optional  *OptHandles
	managers  []HandleManager
}

func NewHandleRegistry() *HandleRegistry {
	registry := &HandleRegistry{
		Alloc:     NewHandleAllocator(StdioMaxID + 1),
		Files:     NewFileHandles(),
		Dirs:      NewDirHandles(),
		Sockets:   NewSocketHandles(),
		Processes: NewProcessHandles(),
		Optional:  NewOptHandles(),
	}

	registry.managers = []HandleManager{
		registry.Files, registry.Dirs, registry.Sockets, registry.Processes, registry.Optional,
	}

	return registry
}

func (r *HandleRegistry) CloseAll() {
	for _, manager := range r.managers {
		manager.CloseAll()
	}
}

func (r *HandleRegistry) TransferTo(dst *HandleRegistry, id int) (int, string) {
	if id >= 0 && id <= StdioMaxID {
		return -1, "Cannot transfer standard handles"
	}

	if r == dst {
		if _, ok := r.Files.Get(id); ok {
			return id, ""
		}

		if _, ok := r.Dirs.Get(id); ok {
			return id, ""
		}

		if _, ok := r.Sockets.Get(id); ok {
			return id, ""
		}

		if _, ok := r.Processes.Get(id); ok {
			return id, ""
		}

		if _, ok := r.Optional.Get(id); ok {
			return id, ""
		}

		return -1, "Invalid handle"
	}

	if file, ok := r.Files.Extract(id); ok {
		r.Alloc.Free(id)
		newID := dst.Alloc.Alloc()
		dst.Files.Store(newID, file)

		return newID, ""
	}

	if dir, ok := r.Dirs.Extract(id); ok {
		r.Alloc.Free(id)
		newID := dst.Alloc.Alloc()
		dst.Dirs.Store(newID, dir)

		return newID, ""
	}

	if socket, ok := r.Sockets.Extract(id); ok {
		r.Alloc.Free(id)
		newID := dst.Alloc.Alloc()
		dst.Sockets.Store(newID, socket)

		return newID, ""
	}

	if proc, ok := r.Processes.Extract(id); ok {
		r.Alloc.Free(id)
		newID := dst.Alloc.Alloc()
		dst.Processes.Store(newID, proc)

		return newID, ""
	}

	if opt, ok := r.Optional.Extract(id); ok {
		r.Alloc.Free(id)
		newID := dst.Alloc.Alloc()
		dst.Optional.Store(newID, opt)

		return newID, ""
	}

	return -1, "Invalid handle"
}
