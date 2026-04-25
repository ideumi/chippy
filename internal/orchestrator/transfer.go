/*
 *
 * RR2 - internal/orchestrator/transfer.go
 *
 */

package orchestrator

import (
	"chip-go/internal/context"
	"chip-go/internal/handles"
)

// Transfer moves a handle from the caller's actor to dstID's registry. Returns
// the new destination handle ID on success, or (-1, reason) on failure.
func (o *Orchestrator) Transfer(srcCtx interface{}, dstID, handleID int) (int, string) {
	srcInstanceID := context.GetInstanceID(srcCtx)

	o.mu.RLock()

	src := o.instances[srcInstanceID]
	dst := o.instances[dstID]

	o.mu.RUnlock()

	if src == nil || dst == nil {
		return -1, "Invalid actor handle"
	}

	return src.Registry.TransferTo(dst.Registry, handleID)
}

func (o *Orchestrator) GetRegistry(ctx interface{}) *handles.HandleRegistry {
	instanceID := context.GetInstanceID(ctx)

	o.mu.RLock()

	inst := o.instances[instanceID]

	o.mu.RUnlock()

	if inst != nil {
		return inst.Registry
	}

	return nil
}

func (o *Orchestrator) GetRR2ForContext(ctx interface{}) RR2Interface {
	instanceID := context.GetInstanceID(ctx)

	o.mu.RLock()

	inst := o.instances[instanceID]

	o.mu.RUnlock()

	if inst != nil {
		return inst.RR
	}

	return nil
}
