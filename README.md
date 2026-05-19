<p align="center">
  <img src="media/ChipLogo.svg" alt="ChipLang Logo" width="200"/>
</p>

<p align="center">
  <a href="https://codeberg.org/ideumi/chip-go/releases"><img src="https://img.shields.io/gitea/v/release/ideumi/chip-go?gitea_url=https%3A%2F%2Fcodeberg.org&label=release" alt="Latest release"/></a>
  <img src="https://img.shields.io/badge/license-BSD--2--Clause-green" alt="License"/>
</p>

# ChipLang

ChipLang (Chipmunk language) is an interpreted scripting / programming language project written in Go, it has the goal of creating a simple, modular, low profile, understandable and hackable programming language for scripting and tooling that would be hard to write, maintain, architect and deploy in shell.

I aim to make ChipLang maintain a clean separation between builtins and libraries where possible. Builtins provide fundamental primitives for common operations that require performance or cannot be implemented cleanly in Chip itself. The native core library (`lib/`) abstracts these primitives for convenience. The codebase is to stay simple and accessible on both the Go and Chip sides.

### Installation for [Supported Platforms](#supported-platforms)

```bash
curl -fsSL https://codeberg.org/ideumi/chip-go/raw/branch/main/scripts/net-install.sh | sh
```

### [Releases](https://codeberg.org/ideumi/chip-go/releases) | [Quick Start](QUICKSTART.md)

## Features

**Language**
- Dynamically typed, imperative (procedural) interpreted programming / scripting language
- Clean, readable syntax with familiar C-like control flow (`if`, `elseif`, `else`, `while`, `for`)
- Simple data types: numbers, strings, bytes, lists and maps
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
- JSON encoding and decoding
- HTTP client
- TLS sockets
- Hashing (SHA-2, SHA-3 etc.)
- ....

**Built-in Primitives**
- 70+ built-in functions for core operations
- Direct file descriptor and socket access
- File system operations (stat, chmod, symlinks, etc.)
- Directory traversal and file metadata
- Environment variable management
- Process control
- Type introspection and conversion
- Low-level string and byte operations
- ....

**Concurrency**
- Actor model
- Message passing
- Operating system signals
- Automatic deadlock detection

**Developer Tools**
- Interactive REPL with history for experimentation
- Syntax files in `ide/` for:
  - KDE
  - GNOME / GTK
  - Vim
  - VSCodium / VSCode
- Built-in documentation system (`chippy doc`)
- Build system (`chippy combine`) with dependency resolution and bundling
    - Syntax validation and symbol collision detection

### Hello World
```chiplang
load("libprint.chh");
Println("Hello World");
```

## Supported Platforms

- Any reasonably modern FHS-compliant Linux distro
- Android (via Termux)

## Supported Architectures

- x86_64
- aarch64

## Installation

ChipLang can be installed by running the following command in your terminal.

```bash
curl -fsSL https://codeberg.org/ideumi/chip-go/raw/branch/main/scripts/net-install.sh | sh
```

or by using the...

### Manual Method

1. Download the tarball matching your architecture from the [releases page](https://codeberg.org/ideumi/chip-go/releases).
2. Extract: `tar -xf chiplang-<version>-linux-<arch>.tar.xz`.
3. Run `sudo ./install.sh` (or `./install.sh` on Termux) from inside the extracted directory.

## Uninstallation

Run `sudo ./uninstall.sh` from the extracted tarball directory (or `./uninstall.sh` on Termux).

## Building from Source

Requirements:
- Go 1.24.4 or later
- git
- make

```bash
git clone https://codeberg.org/ideumi/chip-go
cd chip-go
make
```

## Building the Release Package

```bash
# Build release your current platform
make release    

# Build releases for x86_64 and aarch64
make release-all
```

## Quick Start

Please read the [QUICKSTART](QUICKSTART.md) Guide.

## License

ChipLang is licensed under the 2-Clause BSD License. See `LICENCE.txt`.

Third-party components are licensed under their respective licenses. See `LICENCES_THIRDPARTY.txt` and `thirdparty/`.
