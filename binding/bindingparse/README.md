# bindingparse

This package provides a minimal, reference implementation of parser functions that satisfy the `binding.Parser` interface.

## A Deliberately Flawed Example

The parsers in this package are **not intended for production use**. They are intentionally simple—thin wrappers around the standard `strconv` package that expose underlying errors directly.

Using this package in a real application would be lazy. It exists only to demonstrate the `binding.Parser` interface. For any serious project, you should write your own `parser` package with robust error handling and messages that are meaningful to your application's users.

## Why the Awkward Name?

The name `bindingparse` is intentionally verbose and inconvenient.

It is a signal. A good, application-specific parser package should simply be named `parser`. The cumbersome name of this package is meant to discourage its continued use and to constantly remind you to replace it with your own, properly designed implementation.
