/*
 *
 * Chippy - internal/orchestrator/orchestrator.go
 *
 */

package orchestrator

import (
	"chip-go/internal/handles"
	"sync"
)

type Orchestrator struct {
	instances map[int]*Instance
	nextID    int
	mu        sync.RWMutex
	factory   ModenaFactory
}

var global *Orchestrator

func New() *Orchestrator {
	orch := &Orchestrator{
		instances: make(map[int]*Instance),
		nextID:    1,
	}

	global = orch
	return orch
}

func Get() *Orchestrator {
	return global
}

func (o *Orchestrator) SetFactory(factory ModenaFactory) {
	o.factory = factory
}

func (o *Orchestrator) CreateMain(mod Modena) *Instance {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst := &Instance{
		ID:       0,
		Modena:   mod,
		Registry: handles.NewHandleRegistry(),
		Inbox:    NewInbox(),
	}

	o.instances[0] = inst

	return inst
}

func (o *Orchestrator) CreateActor() *Instance {
	o.mu.Lock()
	defer o.mu.Unlock()

	id := o.nextID
	o.nextID++

	inst := &Instance{
		ID:       id,
		Registry: handles.NewHandleRegistry(),
		Inbox:    NewInbox(),
		ResultCh: make(chan ActorResult, 1),
	}

	o.instances[id] = inst

	return inst
}

func (o *Orchestrator) InitActorModena(inst *Instance) {
	if o.factory != nil {
		inst.Modena = o.factory(inst.ID)
	}
}

func (o *Orchestrator) GetInstance(id int) *Instance {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.instances[id]
}

func (o *Orchestrator) RemoveInstance(id int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.instances, id)
}

// MarkWaited returns false on the second and later calls so wait() can reject
// a double-wait.
func (o *Orchestrator) MarkWaited(inst *Instance) bool {
	o.mu.Lock()
	defer o.mu.Unlock()

	if inst.waited {
		return false
	}

	inst.waited = true

	return true
}

// AddLoadedOpt is idempotent.
func (o *Orchestrator) AddLoadedOpt(inst *Instance, name string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, existing := range inst.loadedOpts {
		if existing == name {
			return
		}
	}

	inst.loadedOpts = append(inst.loadedOpts, name)
}

func (o *Orchestrator) GetLoadedOpts(inst *Instance) []string {
	o.mu.RLock()
	defer o.mu.RUnlock()

	out := make([]string, len(inst.loadedOpts))
	copy(out, inst.loadedOpts)

	return out
}
