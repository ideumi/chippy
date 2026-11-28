# Developer Quick Start Guide

## Introduction to the `combine` build-system

> This section is an explanation of what `combine` is and how it works with the `load(filename)` function. It is a recommended read. However, if you want to start the tutorial right away, go to [Step 1: Project Setup](#step-1-project-setup).

`combine` is Chippy's build-system. Its primary task is to create bundles of all dependencies that your program needs and your program itself. That way you do not need complex scripts to ship your program, you just ship the bundle and it will work if a `chippy` interpreter is installed on the target machine. However, you do not need to use `combine` to use Chippy if you do not want to, or do not need it:

Chippy programs fall broadly into two categories: those that need `combine`, and those that do not.

The dividing line is how `load(filename)` is used. `load(filename)` is resolved against the current directory and has no search functionality, so:

- **A program with no `load(filename)` calls** at all runs directly without building it with `combine`.

- **A program that loads all of its dependencies by pointing `load(filename)` to existing code on disk** will work without `combine`, given all the files that are loaded carry valid `load(filename)`s of their own that point to valid code on disk.

- **A program that loads the core library**, like `load("libprint.chh")`, needs `combine`. This also applies if any code loaded contains `load(filename)` calls that do not point to valid code on disk. The core library lives in `/usr/lib/chippy/` and is designed for `combine`, as many of its modules depend on each other. `combine` searches its configured `Paths` from `combine.chp`, finds and collects everything your program needs, and then directly runs, or bundles your program to an executable.

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

This creates a template `combine.chp` file that is ready to use for this demonstration:

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

## Step 3: Run and Build

You can now run:

```bash
chippy combine run
```

in the same directory where your `combine.chp` lives, to run your program directly without creating a bundle file on disk.

> TIP: Using `chippy combine run` is recommended for debugging as you will get better error messages in case something goes wrong.

You can also create an executable bundle on disk with:

```bash
chippy combine
```

> TIP: `chippy combine` by default looks for a file called `combine.chp`, you can manually specify a combine config file by using `chippy combine <config>`.

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

### Run the bundle

If you created a bundle, you can then execute it with:

```bash
chippy myprogram
```

or

```bash
./myprogram
```

> TIP: If you want to make changes to your program, just make the changes, save the file and run `chippy combine` again, your program will get automatically overwritten.

## Step 4: Try out the Development Tools

Chippy provides developers with some useful tooling:

### Check

This allows you to verify if your code has syntax errors before building:

```bash
chippy check main.chp
```

You can also check every `.chp` and `.chh` file in the project recursively:

```bash
chippy check all
```

### Format

This formats your source files in place to the Chippy style:

```bash
chippy format main.chp
```

You can also check which files need formatting with:

```bash
chippy format check main.chp
```

You can also check every `.chp` and `.chh` file in the project recursively:

```bash
chippy format check all
```

> TIP: These commands are compatible with, and can be added to a build system of your choice such as `make`.

---

## Important Notes

- **Semicolons are required**: Every statement must end with `;`
- **1-based indexing**: Lists and strings start at index 1, not 0 like in other languages
- **Error handling**: `iserr(value)` and `isok(value)` are useful to check results.

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
func Add(a, c)
{
	return a + c;
}

var result = Add(10, 20);
```

### Control Structures
```chippy
if condition
{
	# code
}

if condition
{
	# code
}
elseif otherCondition
{
	# code
}
else
{
	# code
}

for i = 1 to 10
{
	# code
}

for i = 1 to 10 step 2
{
	# code
}

while condition
{
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

Chippy has wide-reaching documentation in appropriate places:

```bash
# List all documentation topics
chippy doc

# View documentation for topic
chippy doc <topic>
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

- `chippy --help`
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
