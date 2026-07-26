/*
 *
 * Modena - internal/values/constant.go
 *
 */

package values

import (
	"math"
	"strconv"
)

// Stricter than == so the whole number 1 and the decimal 1.0 stay apart.
func ConstantIdentity(value Value) string {
	switch typed := value.(type) {
	case *Number:
		if typed.isInt {
			return "i" + strconv.FormatInt(typed.iVal, 10)
		}

		return "f" + strconv.FormatUint(math.Float64bits(typed.fVal), 16)

	case *String:
		return "s" + typed.Value
	}

	panic("ConstantIdentity: unexpected constant type")
}
