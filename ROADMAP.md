# ChipLang Roadmap

## 1.0.x / Current

The language is currently in a state of heavy development where i may add or remove features at will without any advanced notice. Development currently also occurs in an irregular pace, depending on my mood and the time i can find.

I would not recommend writing anything important in Chip as of yet, as there is no stability in regards to the language, the corelib is also as of now still underdeveloped.

Some things i plan to still accomplish in 1.0.x:

- ~~`chippy check` to parse a file for syntax errors~~
  - ~~Integrated in the makefile for `lib/` and `misc/`~~

- ~~Expand `chippy doc` to also be able to be ran on `.chp` and `.chh` files to get an overview via comments.~~
  
- Expand the corelib
  - JSON

- Expand syntax support to more `ide/`s

- Perhaps create a logo or mascot

**Enthusiastic:**

- `chippy format` basically `go fmt` for Chip files.
  - This is probably hell to implement, this feature is dependent on my motivation.

## Version 1.1.x

1.1.0 will be the first release with actual stability, this means:

- Bureaucracy
- A chippy version as well as its plugins are compiled against a fixed Go version
- Additive development and standard warning time for deprecation of anything
  - Deprecation in the standard language is intended to be a rare case and only done if absolutely needed
- Bug and security fixes
- `chippy` will be stable without any breaking changes
- `builtins/` will be stable without any breaking changes
- `lib/` will be stable without any breaking changes

1.1.0 is planned for some time in the summer of 2026.

## Release Names

I like naming releases after various things for fun

| Version |   Name    |        Meaning        |
|---------|-----------| --------------------- |
| 1.0.x   | pardalote | Small Australian birb |
| 1.1.x   |    TBA    |           /           |
| 1.2.x   |    TBA    |           /           |
