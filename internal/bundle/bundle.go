/*
 *
 * RR2 - internal/bundle/bundle.go
 *
 */

package bundle

import (
	"chip-go/internal/constants"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	MagicBytes    = "CHIPBIN"
	FormatVersion = uint8(1)
	ShebangLine   = constants.COMBINE_SHEBANG + "\n"
)

type FileType uint8

const (
	FileTypePlugin FileType = 1
	FileTypeAsset  FileType = 2
)

type Bundle struct {
	// Metadata
	Project         string
	Version         string
	ChipLangVersion string
	BundleID        uint64

	// Platform info
	HasPlugins bool
	TargetOS   string // Only relevant if HasPlugins
	TargetArch string // Only relevant if HasPlugins

	// Content
	Program string
	Files   []BundleFile
}

type BundleFile struct {
	Name string
	Type FileType
	Data []byte
}

func (b *Bundle) DetectPlugins() {
	for _, f := range b.Files {
		if f.Type == FileTypePlugin {
			b.HasPlugins = true
			b.TargetOS = runtime.GOOS
			b.TargetArch = runtime.GOARCH

			return
		}
	}
}

func (b *Bundle) GetPlugins() []BundleFile {
	var plugins []BundleFile

	for _, f := range b.Files {
		if f.Type == FileTypePlugin {
			plugins = append(plugins, f)
		}
	}

	return plugins
}

func (b *Bundle) GetAssets() []BundleFile {
	var assets []BundleFile

	for _, f := range b.Files {
		if f.Type == FileTypeAsset {
			assets = append(assets, f)
		}
	}

	return assets
}

func (b *Bundle) Validate() error {
	if b.HasPlugins {
		if b.TargetOS != runtime.GOOS {
			return fmt.Errorf("Bundle built for %s, running on %s", b.TargetOS, runtime.GOOS)
		}

		if b.TargetArch != runtime.GOARCH {
			return fmt.Errorf("Bundle built for %s, running on %s", b.TargetArch, runtime.GOARCH)
		}

		if b.ChipLangVersion != constants.STR_LPLVR {
			return fmt.Errorf("Plugin bundle requires ChipLang %s, running %s",
				b.ChipLangVersion, constants.STR_LPLVR)
		}
	}

	return nil
}

func ExtractFile(file BundleFile, extractDir string) error {
	// No paths, only names
	if strings.ContainsAny(file.Name, "/\\") || filepath.Base(file.Name) != file.Name {
		return fmt.Errorf("invalid filename in bundle: %s", file.Name)
	}

	filePath := filepath.Join(extractDir, file.Name)

	// Determine permissions based on file type
	perm := os.FileMode(constants.FILE_PERM_READABLE)

	if file.Type == FileTypePlugin {
		perm = constants.FILE_PERM_EXECUTABLE
	}

	if err := os.WriteFile(filePath, file.Data, perm); err != nil {
		return err
	}

	return nil
}

func (b *Bundle) Extract(extractDir string) error {
	for _, file := range b.Files {
		if err := ExtractFile(file, extractDir); err != nil {
			return err
		}
	}

	return nil
}

func (b *Bundle) GetProgram(extractDir string) string {
	// Replace placeholder
	return strings.ReplaceAll(
		b.Program,
		fmt.Sprintf("$CHIPBIN_EXTRACT/%d", b.BundleID),
		extractDir,
	)
}
