/*
 *
 * RR2 - internal/bundle/writer.go
 *
 */

package bundle

import (
	"chip-go/internal/constants"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func WriteBundle(path string, bundle *Bundle) error {
	f, err := os.Create(path)

	if err != nil {
		return fmt.Errorf("creating bundle file: %w", err)
	}

	defer f.Close()

	if _, err := f.Write([]byte(ShebangLine)); err != nil {
		return fmt.Errorf("writing shebang: %w", err)
	}

	if err := writeHeader(f, bundle); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	if err := writeProgram(f, bundle.Program); err != nil {
		return fmt.Errorf("writing program: %w", err)
	}

	if err := writeFiles(f, bundle.Files); err != nil {
		return fmt.Errorf("writing files: %w", err)
	}

	if err := os.Chmod(path, constants.FILE_PERM_EXECUTABLE); err != nil {
		return fmt.Errorf("setting executable permissions: %w", err)
	}

	return nil
}

func writeHeader(w io.Writer, b *Bundle) error {
	if _, err := w.Write([]byte(MagicBytes)); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, FormatVersion); err != nil {
		return err
	}

	if err := writeString(w, b.Project); err != nil {
		return err
	}

	if err := writeString(w, b.Version); err != nil {
		return err
	}

	if err := writeString(w, b.ChipLangVersion); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, b.BundleID); err != nil {
		return err
	}

	hasPlugins := uint8(0)
	if b.HasPlugins {
		hasPlugins = 1
	}

	if err := binary.Write(w, binary.LittleEndian, hasPlugins); err != nil {
		return err
	}

	// only if bundle has plugins
	if b.HasPlugins {
		if err := writeString(w, b.TargetOS); err != nil {
			return err
		}

		if err := writeString(w, b.TargetArch); err != nil {
			return err
		}
	}

	return nil
}

func writeProgram(w io.Writer, program string) error {
	// Program size
	if err := binary.Write(w, binary.LittleEndian, uint64(len(program))); err != nil {
		return err
	}

	// Program content
	if _, err := w.Write([]byte(program)); err != nil {
		return err
	}

	return nil
}

func writeFiles(w io.Writer, files []BundleFile) error {
	// Number of files
	if err := binary.Write(w, binary.LittleEndian, uint32(len(files))); err != nil {
		return err
	}

	// Write file table
	for _, file := range files {
		// Filename
		if err := writeString(w, file.Name); err != nil {
			return err
		}

		// File type
		if err := binary.Write(w, binary.LittleEndian, uint8(file.Type)); err != nil {
			return err
		}

		// File size
		if err := binary.Write(w, binary.LittleEndian, uint64(len(file.Data))); err != nil {
			return err
		}
	}

	// Write file data
	for _, file := range files {
		if _, err := w.Write(file.Data); err != nil {
			return err
		}
	}

	return nil
}

func writeString(w io.Writer, s string) error {
	// String length
	if err := binary.Write(w, binary.LittleEndian, uint32(len(s))); err != nil {
		return err
	}

	// String content
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}

	return nil
}
