<p align="center">
  <img src="media/ChipLogo.svg" alt="Chippy Logo" width="200"/>
</p>

<p align="center">
  <a href="https://codeberg.org/ideumi/chippy/releases"><img src="https://img.shields.io/gitea/v/release/ideumi/chippy?gitea_url=https%3A%2F%2Fcodeberg.org&label=release" alt="Latest release"/></a>
  <img src="https://img.shields.io/badge/license-BSD--2--Clause-green" alt="License"/>
</p>

# Chippy

Chippy is an interpreted scripting and programming language written in Go. It aims to be simple, modular, and hackable: explicit over abstract, imperative, easy to trace, with direct access to the operating system. It's built for tooling that has outgrown the shell.

### Installation for [Supported Platforms](#supported-platforms)

Chippy can be installed or updated to the latest version by running the following command in your terminal:

```bash
curl -fsSL https://codeberg.org/ideumi/chippy/raw/branch/main/scripts/net-install.sh | sh
```
> It is best practice to review any script before running it in your terminal. You can read it [here](scripts/net-install.sh).

### [View Releases](https://codeberg.org/ideumi/chippy/releases) | [Getting Started](#getting-started) | [Developer Quick Start Guide](QUICKSTART.md)

## Features

**Language**
- Dynamically typed, imperative (procedural) interpreted programming / scripting language
- Clean, readable syntax with familiar C-like control flow (`if`, `elseif`, `else`, `while`, `for`)
- Simple data types: numbers, strings, bytes, lists and maps
- Bytecode compilable and powered by a stack virtual machine
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
- Built-in documentation system for language features, builtins and source files (`chippy doc`)
- Built-in formatter (`chippy format`)
- Build system (`chippy combine`) with dependency resolution and bundling
    - Syntax validation and symbol collision detection
- Built-in bytecode compiler (`chippy compile`) and disassembler (`chippy disasm`)

**Hello World Sample**
```chippy
load("libprint.chh");
Println("Hello World");
```

## Supported Platforms

- Any reasonably modern FHS-compliant Linux distro
- Android (via Termux)

## Supported Architectures

- x86_64
- aarch64

## Installation / Updating

Chippy can be installed or updated to the latest version by running the following command in your terminal:

```bash
curl -fsSL https://codeberg.org/ideumi/chippy/raw/branch/main/scripts/net-install.sh | sh
```
> It is best practice to review any script before running it in your terminal. You can read it [here](scripts/net-install.sh).

or by using the...

### Manual Method

1. Download the tarball matching your architecture from the [releases page](https://codeberg.org/ideumi/chippy/releases).
2. Extract: `tar -xf chippy-<version>-linux-<arch>.tar.xz`.
3. Run `sudo ./install.sh` (or `./install.sh` on Termux) from inside the extracted directory.

> If you need an older release, you can download those over at the [GitHub releases page](https://github.com/ideumi/chippy/releases).

## Getting Started

Once installed, the interpreter is `chippy`. Run `chippy --help` for an overview of the available commands.

### For Users

To run a Chippy program, run it either directly with `./<program>` or by using `chippy <program>`.

### For Developers

To start writing your own Chippy programs, read the [Developer Quick Start Guide](QUICKSTART.md).

## Uninstallation

Run `sudo ./uninstall.sh` from the extracted tarball directory (or `./uninstall.sh` on Termux).

## Building from Source

Requirements:
- Go 1.26.5 or later
- git
- make

```bash
git clone https://codeberg.org/ideumi/chippy
cd chippy
make
```

## Building the Release Package

```bash
# Build release for your current platform
make release

# Build releases for x86_64 and aarch64
make release-all
```

## Version History

| Code name      | Series        | Stable API    | Timespan              | Version ranges | Status      | CN refers to
| -------------- | ------------- | ------------- | --------------------- | -------------- | ----------- | ------------ 
| pardalote      | 1.0.x         | No            | Sep. 2025 - Aug. 2026 | 1.0.0 - 1.0.24 | Finished    | [Pardalotes](https://en.wikipedia.org/wiki/Pardalote)
| rixosa         | 1.1.x         | No            | Aug. 2026 - TBD       | 1.1.0 - TBD    | **Current** | [Cattle tyrant](https://en.wikipedia.org/wiki/Cattle_tyrant)
| canaria        | 1.2.x         | TBD           | TBD                   | /              | /           | [Atlantic canary](https://en.wikipedia.org/wiki/Atlantic_canary)
| TBA            | 1.3.x         | TBD           | TBD                   | /              | /           | TBA

For detailed information on what changed between releases see the [changelogs](changelog/).

## License

Chippy is licensed under the 2-Clause BSD License. See `LICENCE.txt`.

Third-party components are licensed under their respective licenses. See `LICENCES_THIRDPARTY.txt` and `thirdparty/`.
