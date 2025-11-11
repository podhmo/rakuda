# rakudamiddleware

This package provides a collection of middleware for the `rakuda` HTTP router.

## Package Naming Conventions

The name `rakudamiddleware` was chosen to be `goimports`-friendly. It is common for developers to create their own packages named `middleware` or `handler` within their application code. If `rakuda` also used these generic names, it could lead to naming collisions, which can be cumbersome to resolve.

To avoid this, we have intentionally prefixed the package name with `rakuda`. This makes the package name more verbose than is typical for Go, but it prevents potential conflicts and makes it clear that the package is part of the `rakuda` ecosystem.

Note: This naming convention is an intentional deviation from the guidelines in [Effective Go](https://go.dev/doc/effective_go#names), which recommend short, concise package names.

### Reference Implementation

The middleware provided in the `rakudamiddleware` package is intended as a reference implementation. It is expected that developers will eventually replace these with their own middleware that is better suited to their specific environment and needs. The verbose package name also serves to emphasize this point.
