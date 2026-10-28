/*
 *
 * Chippy - internal/constants/limits.go
 *
 */

package constants

const (
	LIMIT_VALUE_NESTING_DEPTH = 100000
	LIMIT_CALL_DEPTH          = 10000
	LIMIT_DECODE_DEPTH        = 10000

	LIMIT_REPEAT_SIZE     = 1 << 29
	LIMIT_REPEAT_ELEMENTS = LIMIT_REPEAT_SIZE / 32 // 32 bytes on 64-bit, keep list results at ~LIMIT_REPEAT_SIZE
)
