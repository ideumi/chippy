/*
 *
 * Modena - internal/bytecode/deduper.go
 *
 */

package bytecode

type deduper[K comparable] struct {
	slotOf map[K]uint32
	count  uint32
}

func (d *deduper[K]) add(value K) (slot uint32, firstTime bool) {
	if slot, seen := d.slotOf[value]; seen {
		return slot, false
	}

	if d.slotOf == nil {
		d.slotOf = make(map[K]uint32)
	}

	slot = d.count
	d.slotOf[value] = slot
	d.count++

	return slot, true
}

func (d *deduper[K]) dropLookup() {
	d.slotOf = nil
}
