> [!NOTE]
> This feature has been implemented.

# Plan: SSE (Server-Sent Events) Responder

This document outlines the plan to add Server-Sent Events (SSE) support to the `rakuda` router.

## 1. Goal

The primary goal is to create a new `SSE` method on the `Responder` struct that simplifies the process of sending SSE streams to clients. This is particularly useful for applications like real-time updates, notifications, and streaming responses from LLMs.

The implementation should feel familiar to developers already using the `JSON` method.

## 2. Design

### `SSE` Function

A new standalone generic function will be added in `responder.go`.

```go
// SSE streams data from a channel to the client using the Server-Sent Events protocol.
// It sets the appropriate headers and handles the event stream formatting.
func SSE[T any](responder *Responder, w http.ResponseWriter, req *http.Request, ch <-chan T)
```

**Note on Design:** While the `JSON` method is part of the `Responder` struct, `SSE` is implemented as a standalone function. This is due to a current limitation in Go (as of Go 1.22), which does not permit methods to have their own type parameters. A standalone generic function is the idiomatic way to achieve this functionality while still leveraging the `Responder`'s configuration (e.g., for logging).

**Parameters:**

-   `w http.ResponseWriter`: The HTTP response writer.
-   `req *http.Request`: The HTTP request.
-   `ch <-chan T`: A read-only channel from which the data to be sent is received. The generic type `T` allows for streaming any kind of data that can be marshaled.

### Behavior

1.  **Set Headers**: The method will first set the required SSE headers:
    -   `Content-Type: text/event-stream`
    -   `Cache-Control: no-cache`
    -   `Connection: keep-alive`
2.  **Flush Support**: It will check if the `http.ResponseWriter` supports flushing (`http.Flusher`) to ensure that data is sent to the client immediately.
3.  **Event Loop**: The method will enter a loop that listens for two main events:
    -   Data arriving on the `ch` channel.
    -   The client closing the connection (detected via `req.Context().Done()`).
4.  **Data Formatting**: Each item received from the channel will be formatted as a standard SSE `data` message. For non-string types, it will be JSON-marshaled.
    ```
    data: {"message": "hello"}

    ```
5.  **Termination**: The loop will terminate when either the channel is closed or the client disconnects.

### Error Handling

-   If `json.Marshal` fails for a message, the error will be logged using the logger from the request context (or the `DefaultLogger`), and the connection will be kept alive to process subsequent messages.
-   The method will gracefully handle client disconnects.

## 3. Implementation Steps

1.  **Modify `responder.go`**: Add the `SSE` method to the `Responder` struct with the logic described above.
2.  **Add `*Event` helper**: Add a helper for sending events with custom names.
3.  **Modify `responder_test.go`**: Create a new test case, `TestResponder_SSE`, to verify the functionality.

## 4. Test Plan

The test case in `responder_test.go` will:

1.  Use `httptest.NewRecorder` to capture the response.
2.  Create a channel and send a few sample data structures to it.
3.  Call `responder.SSE()` in a separate goroutine.
4.  Read the response body and verify that:
    -   The headers are set correctly.
    -   The data is formatted as valid SSE messages.
    -   The connection is held open until the channel is closed.
5.  Test with different data types (structs, strings, etc.).
6.  Test the client disconnect scenario.

This plan ensures a robust and well-tested implementation that integrates seamlessly into the existing `rakuda` framework.
