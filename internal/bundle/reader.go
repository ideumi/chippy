/*
 *
 * RR2 - internal/bundle/reader.go
 *
 */

package bundle

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func ReadBundle(path string) (*Bundle, error) {
	f, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("opening bundle file: %w", err)
	}

	defer f.Close()

	// Skip shebang
	if _, err := f.Seek(int64(len(ShebangLine)), io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking past shebang: %w", err)
	}

	// Read and validate magic
	magic := make([]byte, len(MagicBytes))

	if _, err := f.Read(magic); err != nil {
		return nil, fmt.Errorf("reading magic bytes: %w", err)
	}

	if string(magic) != MagicBytes {
		return nil, fmt.Errorf("not a valid bundle")
	}

	bundle := &Bundle{}

	// Read header
	if err := readHeader(f, bundle); err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	// Read program
	if err := readProgram(f, bundle); err != nil {
		return nil, fmt.Errorf("reading program: %w", err)
	}

	// Read files
	if err := readFiles(f, bundle); err != nil {
		return nil, fmt.Errorf("reading files: %w", err)
	}

	return bundle, nil
}

func readHeader(r io.Reader, b *Bundle) error {
	// Format version
	var formatVersion uint8

	if err := binary.Read(r, binary.LittleEndian, &formatVersion); err != nil {
		return err
	}

	if formatVersion != FormatVersion {
		return fmt.Errorf("unsupported bundle format version %d (expected %d)", formatVersion, FormatVersion)
	}

	// Project name
	project, err := readString(r)

	if err != nil {
		return err
	}

	b.Project = project

	// Version
	version, err := readString(r)

	if err != nil {
		return err
	}

	b.Version = version

	// ChipLang version
	chipVersion, err := readString(r)

	if err != nil {
		return err
	}

	b.ChipLangVersion = chipVersion

	// Bundle ID
	if err := binary.Read(r, binary.LittleEndian, &b.BundleID); err != nil {
		return err
	}

	// Has plugins flag
	var hasPlugins uint8

	if err := binary.Read(r, binary.LittleEndian, &hasPlugins); err != nil {
		return err
	}

	b.HasPlugins = hasPlugins == 1

	// only if bundle has plugins
	if b.HasPlugins {
		targetOS, err := readString(r)

		if err != nil {
			return err
		}

		b.TargetOS = targetOS

		targetArch, err := readString(r)

		if err != nil {
			return err
		}

		b.TargetArch = targetArch
	}

	return nil
}

func readProgram(r io.Reader, b *Bundle) error {
	// Program size
	var programSize uint64

	if err := binary.Read(r, binary.LittleEndian, &programSize); err != nil {
		return err
	}

	// Program content
	programData := make([]byte, programSize)

	if _, err := io.ReadFull(r, programData); err != nil {
		return err
	}
	b.Program = string(programData)

	return nil
}

func readFiles(r io.Reader, b *Bundle) error {
	// Number of files
	var numFiles uint32

	if err := binary.Read(r, binary.LittleEndian, &numFiles); err != nil {
		return err
	}

	// Read file table entries
	type fileEntry struct {
		name  string
		ftype FileType
		size  uint64
	}

	entries := make([]fileEntry, numFiles)

	for i := uint32(0); i < numFiles; i++ {
		// Filename
		name, err := readString(r)

		if err != nil {
			return err
		}

		// File type
		var ftype uint8

		if err := binary.Read(r, binary.LittleEndian, &ftype); err != nil {
			return err
		}

		// Validate file type
		if ftype != uint8(FileTypePlugin) && ftype != uint8(FileTypeAsset) {
			return fmt.Errorf("invalid file type %d for file '%s'", ftype, name)
		}

		// File size
		var size uint64

		if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
			return err
		}

		entries[i] = fileEntry{
			name:  name,
			ftype: FileType(ftype),
			size:  size,
		}
	}

	// Read file data
	for _, entry := range entries {
		data := make([]byte, entry.size)

		if _, err := io.ReadFull(r, data); err != nil {
			return err
		}

		b.Files = append(b.Files, BundleFile{
			Name: entry.name,
			Type: entry.ftype,
			Data: data,
		})
	}

	return nil
}

func readString(r io.Reader) (string, error) {
	var length uint32

	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	data := make([]byte, length)

	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}

	return string(data), nil
}
