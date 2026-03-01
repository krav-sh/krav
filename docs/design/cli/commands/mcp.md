# MCP

Starts the ARCI MCP server for Claude Code integration.

## Synopsis

```text
arci mcp [options]
```

## Description

The MCP server exposes ARCI diagnostic and introspection tools to Claude Code via the Model Context Protocol. It communicates over stdin/stdout using MCP's stdio transport and exits when the client disconnects.

The MCP server acts as a stateless proxy that connects to the running `arci server`'s HTTP API. It discovers the server by reading `.arci/server.json` from the project root (see [server discovery](../../server/discovery.md)). If the `arci server` is not running, tool calls return MCP error responses rather than failing silently.

This command is typically not run manually. Claude Code's MCP settings configure it so that Claude Code launches it automatically.

## Flags

`--project-dir <path>`: project root directory. Default: auto-detected by walking up from cwd. Also settable via `ARCI_PROJECT_DIR`.

## Claude Code configuration

Add to your Claude Code MCP configuration:

```json
{
  "mcpServers": {
    "arci": {
      "command": "arci",
      "args": ["mcp"],
      "cwd": "/path/to/project"
    }
  }
}
```

If Claude Code's working directory is already the project root, you can omit `cwd`.

## See also

- [MCP server design](../../mcp/index.md)
- [Server command](server.md)
