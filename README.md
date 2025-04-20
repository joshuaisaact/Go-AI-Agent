# Go AI Agent Foundation

This project provides a basic foundation for building AI agents in Go that can interact with users and utilize tools. It uses the Anthropic API (Claude) by default.

## Project Structure

- `cmd/agent/main.go`: Main application entry point.
- `pkg/agent/`: Contains the core agent logic (`agent.go`, `inference.go`).
- `pkg/tools/`: Contains tool definitions (`tools.go`, `schema.go`) and implementations.
- `go.mod`, `go.sum`: Go module files.

## Setup

1.  **Install Go**: Ensure you have Go installed (version 1.21 or later).
2.  **Dependencies**: Run `go mod tidy` to install dependencies.
3.  **API Key**: Set the `ANTHROPIC_API_KEY` environment variable with your Anthropic API key:
    ```bash
    export ANTHROPIC_API_KEY='your-api-key-here'
    ```

## Running the Agent

```bash
go run cmd/agent/main.go
```

The agent will start, and you can interact with it in the terminal. Use Ctrl+C to exit.

## Tools

The agent currently supports the following tools:

- `read_file`: Reads the content of a file.
- `list_files`: Lists files and directories in a path.
- `edit_file`: Replaces a string in a file (use with caution!).