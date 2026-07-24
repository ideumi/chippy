/*
 *
 * Modena - internal/bytecode/fault.go
 *
 */

package bytecode

import "fmt"

func ModenaError(format string, args ...any) error {
	return fmt.Errorf("modena: "+format, args...)
}

func ModenaPanic(format string, args ...any) {
	panic(ModenaError(format, args...))
}
