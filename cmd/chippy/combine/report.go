/*
 *
 * Combine Error Reporting Logic
 *
 */

package combine

import (
	"fmt"
	"strings"
)

func parseErrorsText(errs []ValidationError) string {
	var block strings.Builder

	block.WriteString("Errors detected:\n")

	for _, verr := range errs {
		block.WriteString(verr.Error.Error())
		block.WriteString("\n")
	}

	return strings.TrimSuffix(block.String(), "\n")
}

func reportParseErrors(errs []ValidationError) {
	fmt.Println(parseErrorsText(errs))
}

func reportCollisions(collisions []SymbolCollision) {
	fmt.Println(collisionsText(collisions))
}

func collisionsText(collisions []SymbolCollision) string {
	var block strings.Builder

	block.WriteString("Symbol collisions detected:\n")

	for _, collision := range collisions {
		// Handle builtin collisions differently

		if collision.IsBuiltinCollision() {
			block.WriteString(fmt.Sprintf("\n  %s '%s' collides with %s:\n", collision.Locations[0].SymType, collision.Name, collision.SymType))
		} else if collision.IsSameFile() {
			block.WriteString(fmt.Sprintf("\n  %s '%s' defined multiple times in %s:\n", collision.SymType, collision.Name, collision.GetFirstFile()))
		} else {
			block.WriteString(fmt.Sprintf("\n  %s '%s' defined in multiple files:\n", collision.SymType, collision.Name))
		}

		for _, loc := range collision.Locations {
			block.WriteString(fmt.Sprintf("    - %s:%d\n", loc.File, loc.Line))
		}
	}

	return strings.TrimSuffix(block.String(), "\n")
}
