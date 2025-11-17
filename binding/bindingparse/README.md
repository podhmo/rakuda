# bindingparse

This package provides a reference implementation of parser functions that satisfy the `binding.Parser` interface.

## Package Naming Conventions

The name `bindingparse` was chosen to be `goimports`-friendly. It is common for developers to create their own packages named `parser` within their application code. If `rakuda` also used this generic name, it could lead to naming collisions, which can be cumbersome to resolve.

To avoid this, we have intentionally prefixed the package name with `binding`. This makes the package name more verbose than is typical for Go, but it prevents potential conflicts and makes it clear that the package is part of the `rakuda` ecosystem.

Note: This naming convention is an intentional deviation from the guidelines in [Effective Go](httpss://go.dev/doc/effective_go#names), which recommend short, concise package names.
