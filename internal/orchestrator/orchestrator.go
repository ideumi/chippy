/*
 *
 * RR2 - internal/orchestrator/orchestrator.go
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
	factory   RR2Factory
}

var global *Orchestrator

func New() *Orchestrator {
	o := &Orchestrator{
		instances: make(map[int]*Instance),
		nextID:    1,
	}

	global = o
	return o
}

func Get() *Orchestrator {
	return global
}

func (o *Orchestrator) SetFactory(factory RR2Factory) {
	o.factory = factory
}

func (o *Orchestrator) CreateMain(rr RR2Interface) *Instance {
	o.mu.Lock()
	defer o.mu.Unlock()

	inst := &Instance{
		ID:       0,
		RR:       rr,
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

func (o *Orchestrator) InitActorRR2(inst *Instance) {
	if o.factory != nil {
		inst.RR = o.factory(inst.ID)
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
