# Simple REST API Example

This example demonstrates the basic usage of the `rakuda` package to create a simple REST API.

## Running the example

To run the server, execute the following command from the root of the repository:

```bash
go run examples/simple-rest-api/main.go
```

The server will start on port `8080`.

## Endpoints

You can test the endpoints using `curl`:

- **GET /**

  ```bash
  curl http://localhost:8080/
  ```

  **Response:**
  ```json
  {"message":"hello world"}
  ```

- **GET /hello/{name}**

  ```bash
  curl http://localhost:8080/hello/Jules
  ```

  **Response:**
  ```json
  {"message":"hello Jules"}
  ```
