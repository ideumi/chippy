# ChipLang

ChipLang (Chipmunk language) is an interpreted scripting / programming language project written in Go, it has the goal of creating a simple, modular, low profile, understandable and hackable programming language for UNIX system scripting and tooling that would be hard to write, maintain, architect and deploy in shell.

I aim to make ChipLang maintain a clean separation between builtins and libraries where possible. Builtins provide fundamental primitives for common operations that require performance or cannot be implemented cleanly in Chip itself. The native core library (`lib/`) abstracts these primitives for convenience. The codebase is to stay simple and accessible on both the Go and Chip sides.

### Hello World
```chiplang
load("libprint.chh");
Println("Hello World");
```

## Features

**Language**
- Dynamically typed, imperative (procedural) interpreted programming / scripting language
- Clean, readable syntax with familiar C-like control flow (`if`, `elif`, `else`, `while`, `for`)
- Simple data types: numbers, strings, bytes and lists
- UTF-8 native
- Functions with proper scoping
- Very fast startup time

**Standard Library**
- File IO operations
- String manipulation and formatting
- Mathematical functions and constants
- Time and calendar operations
- Path manipulation
- Terminal control
- ....

**Built-in Primitives**
- 60+ built-in functions for core operations
- Direct file descriptor and socket access
- File system operations (stat, chmod, symlinks, etc.)
- Directory traversal and file metadata
- Environment variable management
- Process control and signal handling
- System information and process control
- Type introspection and conversion
- Low-level string and byte operations
- ....

**Developer Tools**
- Interactive REPL with history for experimentation
- Syntax files for KDE, GNOME / GTK and Vim in `ide/`
- Built-in documentation system (`chippy doc`)
- Build system (`chippy combine`) with dependency resolution and bundling
    - Syntax validation and symbol collision detection

## Supported Platforms

- Any reasonably modern FHS-compliant Linux distro

## Experimental Platforms

- OpenBSD and FreeBSD - highly experimental for now and not guaranteed to work
- Android (via Termux)

## Installation

1. Build the release package: `make release` (see "Building from Source" below)
2. Extract the generated archive from the `rel/` directory
3. Run the installer: `install.sh` with elevated permissions

## Uninstallation

1. Run the uninstaller: `uninstall.sh` with elevated permissions

## Building from Source

Requirements:
- Go 1.24.0 or later
- make

```bash
make
```

## Building the Release Package

```bash
make release
```

## Quick Start

```bash
# Start the REPL
chippy

# Run commands via shell
chippy -r "fwrite(pack(\"Hello World\") + b[10], 1);"
chippy --run "fwrite(pack(\"Hello World\") + b[10], 1);"

# View available documentation
chippy doc list

# View build-system documentation
chippy doc combine

# View documentation for specific topic
chippy doc <topic>
```

I recommend having a look at `installer`, `misc` and to a lesser extent `lib` for some examples on how Chip is used in practice.

## License

ChipLang is licensed under the 2-Clause BSD License. See `LICENCE.txt`.

Third-party components are licensed under their respective licenses. See `LICENCES_THIRDPARTY.txt` and `thirdparty/`.

---

> Zwei mal drei macht vier,
widewidewitt und drei macht neune,
ich mach mir die Welt,
widewide wie sie mir gefällt.
- Hey, Pippi Langstrumpf (1969)
