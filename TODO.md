# TODO

> **Note on updating this file:**
> - Do not move individual tasks to the "Implemented" section.
> - A whole feature section (e.g., "Core Router Implementation") should only be moved to "Implemented" when all of its sub-tasks are complete.
> - For partially completed features, use checkboxes (`[x]` for complete, `[-]` for partially complete). A feature is considered partially complete if it has been implemented but has associated tests that are currently disabled.
> - For partially completed features, use checkboxes (`[x]`) to mark completed sub-tasks.

This file tracks implemented features and immediate, concrete tasks.

For more ambitious, long-term features, see [docs/router-design.md](./docs/router-design.md).

## Implemented

- **Project Structure**: Initial repository setup with LICENSE, README.md, and AGENTS.md.
- **Documentation**: Design document outlining the core architecture and philosophy ([docs/router-design.md](./docs/router-design.md)).
- **Responder Package**: Implemented the core `Responder` component for handling JSON and error responses, later refactored into the main `rakuda` package.
- **Lift Function**: Implemented the `lift` function to adapt action handlers (`func(req) (data, error)`) into standard `http.Handler`, supporting custom status codes.
- **rakudatest Test Helper**: Created the `rakudatest` package with a generic `Do` function to simplify handler testing, including status validation and response decoding. ([sketch/plan-rakudatest.md](./sketch/plan-rakudatest.md))
- **Binding Package**: Implemented a type-safe, reflect-free data binding package for request values, including support for form data. ([sketch/plan-binding.md](./sketch/plan-binding.md), [sketch/plan-binding-form-values.md](./sketch/plan-binding-form-values.md))
- **Test Logger**: Integrated a `t.Logf`-based `slog.Handler` into `rakudatest` to automatically inject a test-aware logger into request contexts. ([sketch/plan-test-logger.md](./sketch/plan-test-logger.md))
- **Structured Error Responses for Binding**: Enhanced the `binding` and `responder` packages to produce detailed, structured JSON error responses for validation failures. ([sketch/plan-binding-join.md](./sketch/plan-binding-join.md))
- **Centralized Logging**: Implemented a centralized logging strategy using functional options on the `Builder` and a context-based logger propagation middleware. ([sketch/plan-sharing-logger.md](./sketch/plan-sharing-logger.md))
- **SSE Responder**: Added a responder for streaming data using Server-Sent Events (SSE). It supports `text/event-stream` responses and includes helpers for sending named events. See ([sketch/plan-sse-responder.md](./sketch/plan-sse-responder.md)) for details.
- **CLI**: Added a `proutes` utility to display registered handlers via a command-line flag in the example application.
- **Core Router Implementation**: Implemented a type-safe router using a builder pattern. It supports features like middleware, route grouping, custom "Not Found" handlers, and configurable route conflict handling.

## To Be Implemented

### Example Applications
- [x] **Simple REST API example**: Demonstrate basic usage
- [x] **Middleware demonstration**: Show global and scoped middleware
- [ ] **Nested groups example**: Show route grouping patterns
