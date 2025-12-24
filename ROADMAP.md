# ChipLang Roadmap

## Lay of the Land

ChipLang is a personal, open-source side-project. All development is performed and financed by a single maintainer on a volunteer basis. The direction, timing, and scope of work are therefore guided by personal interest and availability rather than a formal schedule. Users are encouraged to explore the language, but should keep in mind that there is no formal guarantee of long-term stability, security-patch prioritisation, backward compatibility or anything actually being correct.

## 1.0.x / Current

The language is currently in a state of heavy development where i may add or remove features at will without any advanced notice. Development currently also occurs in an irregular pace, depending on my mood and the time i can find.

I would not recommend writing anything important in Chip as of yet, as there is no stability or predictability in regards to the language, the corelib is also as of now still underdeveloped.

Some things i plan to still accomplish in 1.0.x:

- ~~`chippy check` to parse a file for syntax errors~~
  - ~~Integrated in the makefile for `lib/` and `misc/`~~

- ~~Expand `chippy doc` to also be able to be ran on `.chp` and `.chh` files to get an overview via comments.~~

- ~~Create the infrastructure to maintain optional built-ins outside the `builtins/` language core.~~
  - ~~JSON~~

- Expand the corelib
  - ~~JSON~~
  - ....

- Create a CI pipeline for automatic release generation (if Codeberg supports it, i haven't researched this yet)

- Expand syntax support to more `ide/`s
  - Jetbrains?
  - VSCode?

- ~~Create a logo or mascot~~

## Version 1.1.x

1.1.0 will be the first release where i plan to make the language predictable, this means:

- Some level of minimal bureaucracy to plan releases and deprecations
- I will begin to use the releases feature with automated builds (if possible) to provide prebuilt binary packages for supported platforms
- Additive development and standard warning time for deprecation of anything
  - Deprecation in the standard language is intended to be a rare case and only done if absolutely needed
- Bug and security fixes

1.1.0 is planned for some time in the summer of 2026.

## What won't be happening for now

- Windows support
  - Not in the spirit of the project
  
- GUI bindings
  - Technically possible with optionals (e.g. Fyne, Qt)
  - Not feasible as a long term project for me
  - Very large subproject

- FFI
  - Too fragile and complicated for the spirit of the language
  
- Bytecode
  - This will probably happen eventually but not in 1.1.x

## Release Names

I like naming releases after various things for fun

| Version |   Name    |        Meaning        |
|---------|-----------| --------------------- |
| 1.0.x   | pardalote | Small Australian birb |
| 1.1.x   |    TBA    |           /           |
| 1.2.x   |    TBA    |           /           |
