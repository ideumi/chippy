# Developer Quick Start Guide

Chippy programs can generally be categorized in two types: those that use `load("libprint.chh")` (without a path) to load the core library, and those that don't use library dependencies or specify paths manually like `load("lib/libxxx.chh")`. Programs that load the core library need the `combine` tool to bundle dependencies. Programs without library dependencies can run directly.

### What is `combine`:

`combine` is Chippy's "build"-system, its primary task is to create bundles of all dependencies that your program needs and your program itself. That way you do not need complex scripts to ship your program, you just ship the bundle and it will work if a `chippy` interpreter is installed on the target machine.

See `chippy doc combine` for more information.

For this quick demo we'll be using `combine`:

## Step 1: Project Setup

Create a project directory:

```bash
mkdir myproject
cd myproject
```

Create the combine template `combine.chp`:

```bash
chippy combine new
```

This creates a template combine file that is ready to use for this demonstration. e.g.

```chippy
...
# Metadata

var Project = "myprogram";
var Version = "1.0.0";
var Licence = "licence.txt";
var Output = "myprogram";

# Entry point

var Source = "main.chp";
...
```

## Step 2: Write Your Program

Create a file called `main.chp`, feel free to copy this example:

```chippy
#!/usr/bin/chippy

load("libprint.chh");
Println("Hello World");
```

## Step 3: Build

Run combine to bundle your program:

```bash
chippy combine
```

> TIP: `chippy combine` by default looks for a file called `combine.chp`, you can manually specify a combine config file by using `chippy combine [config]`.

When you run `combine` you should see something like this:

```
Bundling myprogram V-1.0.0:
Output: myprogram

Files bundled:
  1. /usr/lib/chippy/libhandles.chh
  2. /usr/lib/chippy/libstrings.chh
  3. /usr/lib/chippy/libprint.chh
  4. main.chp

Total: 4 source(s)
Successfully created myprogram

```

This will have created a program called `myprogram` in the development directory.

## Step 4: Run

You can then execute the program

```bash
chippy myprogram
```

or

```bash
./myprogram
```

Output:
```
Hello World
```

> TIP: If you want to make changes to your program, just make the changes, save the file and run `chippy combine` again, your program will get automatically overwritten.

---

## Important Notes

- **Semicolons are required**: Every statement must end with `;`
- **1-based indexing**: Lists and strings start at index 1, not 0 like in other languages
- **Error handling**: `iserr()` and `isok()` are useful to check results.

## Quick Reference

### Types
```chippy
var text = "text";                                             # String
var number = 42;                                               # Integer or floating point number
var numCollection = [1, 2, 3];                                 # List
var bytesList = b[65, 66, 67];                                 # Bytes
var map = m["key1": "value", "key2": 2, "key3": [1, 2, 3]];    # Map
```

### Functions
```chippy
func Add(a, b) {
    return a + b;
}

var result = Add(10, 20);
```

### Control Structures
```chippy
if condition {
    # code
}

if condition {
    # code
} elseif otherCondition {
    # code
} else {
    # code
}

for i = 1 to 10 {
    # code
}

for i = 1 to 10 step 2 {
    # code
}

while condition {
    # code
}
```

### Operators
```
+  -  *  /  %  ^        # Arithmetic
== != <  >  <= >=       # Comparison
and  or  not  xor       # Logical
band  bor  bnot  bxor   # Bitwise
<<  >>                  # Shift
=                       # Assignment
```

## Core library

Chippy has a corelib that is located in `/usr/lib/chippy/`.
```bash
ls /usr/lib/chippy/
```

## Documentation System

Chippy has wide reaching documentation in appropriate places:

```bash
# List all documentation topics
chippy doc list

# View documentation for topic
chippy doc [topic]
```

Chippy's documentation system also allows you to document code and access it via `chippy doc`.

```bash
# Show all documented symbols of the libprint.chh library
chippy doc /usr/lib/chippy/libprint.chh

# Display documentation for the Print(text) symbol.
chippy doc /usr/lib/chippy/libprint.chh "Print(text)"
```

Generally, all symbols in the corelib are documented this way.

For more information on the corelib, see `chippy doc corelib`

### Recommended reading

- `chippy doc combine`
- `chippy doc corelib`
- `chippy doc builtins`
- `chippy doc constants`
- `chippy doc iserr`
- `chippy doc isok`
- `chippy doc scoping`
- `chippy doc mutability`
- `chippy doc listops`
- `chippy doc operators`
