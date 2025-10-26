# ChipLang

ChipLang (Chipmunk language) is an interpreted scripting / programming language written in Go, it has the goal of creating a simple, low profile, understandable and hackable programming language.

ChipLang's ideology is to have a clean separation between fundamental interpreter primitives (`builtins`) which provide complex core functionality that cannot be implemented cleanly or with the required performance otherwise, and the native core library (`lib`) which abstracts these primitives for convenience and more, and then have Chip programs be mostly independent (except for the `chippy` interpreter) bundles created via an intelligent and easy to use bundling system (`combine`).

Chip's Interpreter is called `chippy`, from `chip` -> `chipi` (Chip Interpreter) -> `chippy`

### Hello World
```chiplang
load("libprint.chh");
Println("Hello, World!");
```

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
go mod tidy
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

## Why

Chip is a fun, private little project I have chosen to undertake myself over the last few months to learn a bit more about language and systems design and to fix several design issues with Chip's evil cousin Lyra while implementing some new stuff that I personally would love to see more of in modern programming / scripting languages, although I don't recommend anyone study this in any way to learn how to implement an optimized interpreter or primitives, because this isn't that. While fairly advanced, there are probably issues in this that my brain doesn't even conceptualize as problems. This is my little sandbox where I play around a bit with programming and write my own little tools, that's enough for now x)

## License

ChipLang is licensed under the 2-Clause BSD License. See `LICENCE.txt`.

Third-party components are licensed under their respective licenses. See `LICENCES_THIRDPARTY.txt`.

---

> Zwei mal drei macht vier,
widewidewitt und drei macht neune,
ich mach mir die Welt,
widewide wie sie mir gefällt.
- Hey, Pippi Langstrumpf (1969)
