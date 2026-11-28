/*
 *
 * Shared helpers
 *
 */

package main

import (
	"os"
	"path/filepath"
	"strings"
)

// walkChippyFiles collects every .chp and .chh file under dir in lexical order.
func walkChippyFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".chp") || strings.HasSuffix(path, ".chh")) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
