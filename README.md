# dumbwrap

[![Go](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**dumbwrap** is a static analysis tool for Go that detects functions and methods whose only purpose is to forward arguments to another function.

These forwarding functions can introduce unnecessary indirection, making code harder to navigate, understand, and maintain without adding meaningful behavior.

## Why dumbwrap?

Consider the following Go code:

```go
func GetUser(id string) (*User, error) {
    return FindUser(id)
}
```

The `GetUser` function does nothing beyond forwarding its arguments to `FindUser`.

While wrappers can be useful for abstraction, compatibility, and API design, unnecessary forwarding functions can introduce:

- **Extra indirection:** More code to navigate when tracing execution.
- **Maintenance overhead:** Additional functions to update as APIs evolve.
- **Reduced readability:** More layers of abstraction without additional behavior.
- **Code complexity:** Unnecessary functions that make the codebase harder to understand.

`dumbwrap` helps identify these patterns so you can decide whether the wrappers are actually needed.

## Installation

Install `dumbwrap` using Go:

```bash
go install github.com/angelokurtis/dumbwrap@latest
```

Make sure your Go binary directory (`$(go env GOPATH)/bin`, or `GOBIN` if configured) is included in your `PATH`.

Verify the installation:

```bash
dumbwrap -h
```

## Usage

```bash
dumbwrap [-flag] [package]
```

Run `dumbwrap` against a Go package to identify functions that only forward arguments to another function.

### Basic examples

Analyze the current package:

```bash
dumbwrap
```

Analyze a specific package:

```bash
dumbwrap ./mypackage
```

Ignore exported forwarding functions:

```bash
dumbwrap -ignore-exported ./mypackage
```

Ignore forwarding methods:

```bash
dumbwrap -ignore-methods ./mypackage
```

Include generated Go files:

```bash
dumbwrap -ignore-generated=false ./mypackage
```

Combine multiple options:

```bash
dumbwrap \
  -ignore-exported \
  -ignore-methods \
  ./mypackage
```

## Options

| Flag                        | Description                                                                                             | Default |
|-----------------------------|---------------------------------------------------------------------------------------------------------|---------|
| `-ignore-exported`          | Ignore exported forwarding functions.                                                                   | `false` |
| `-ignore-generated`         | Ignore generated Go files.                                                                              | `true`  |
| `-ignore-methods`           | Ignore forwarding methods.                                                                              | `false` |
| `-ignore-sig-regexps value` | Ignore calls whose target function signature matches the specified regular expression. Can be repeated. | None    |

### Ignoring exported functions

Exported wrappers may intentionally provide a stable public API or hide implementation details.

To exclude them from the analysis:

```bash
dumbwrap -ignore-exported ./mypackage
```

### Ignoring generated code

Generated Go files are ignored by default to avoid reporting wrappers produced by code generators.

To include them in the analysis:

```bash
dumbwrap -ignore-generated=false ./mypackage
```

### Ignoring methods

Some methods exist primarily to satisfy interfaces or delegate behavior to another implementation.

To exclude forwarding methods:

```bash
dumbwrap -ignore-methods ./mypackage
```

### Ignoring function signatures

The `-ignore-sig-regexps` flag allows you to filter forwarding functions based on the signature of the function being called.

This is particularly useful when certain forwarding patterns are intentional, such as logging, tracing, or delegation layers.

For example:

```bash
dumbwrap \
  -ignore-sig-regexps '.*Logger.*' \
  ./mypackage
```

The flag can be repeated to ignore multiple signature patterns:

```bash
dumbwrap \
  -ignore-sig-regexps '.*Logger.*' \
  -ignore-sig-regexps '.*Tracer.*' \
  ./mypackage
```

## Contributing

Contributions are welcome!

If you find a bug, have a feature request, or want to improve the project, feel free to [open an issue](https://github.com/angelokurtis/dumbwrap/issues) or submit a [pull request](https://github.com/angelokurtis/dumbwrap/pulls).

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/angelokurtis/dumbwrap/blob/main/LICENSE) file for details.