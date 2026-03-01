# Dashboard

The arci dashboard is a web-based interface for monitoring and debugging hook activity. It provides real-time visibility into hook events, rule matching, configuration status, and state store contents.

The dashboard is served by the [daemon](../daemon.md) on the same HTTP port as the API. It is designed for developers debugging their hook configurations, answering questions like: Why did this rule fire? Why didn't this rule fire? What's the current state? Is my configuration valid?

## Documentation

- [Features](features.md) — the five user-facing views and what they display
- [Technology](technology.md) — stack choices (Go templates, htmx, Alpine.js, Pico CSS)
- [Real-time updates](real-time.md) — WebSocket streaming, polling, htmx integration
- [Routing](routing.md) — chi router, handler pattern, daemon data flow

## Routes

| Method | Path                | Description                                      |
|--------|---------------------|--------------------------------------------------|
| GET    | `/dashboard/`       | Main page with live event stream and navigation  |
| GET    | `/dashboard/events` | Live event stream, embeddable as htmx fragment   |
| GET    | `/dashboard/stats`  | Rule match statistics with sorting and filtering |
| GET    | `/dashboard/config` | Configuration sources, validation, merged rules  |
| GET    | `/dashboard/state`  | State store browser with filtering and editing   |
| GET    | `/dashboard/test`   | Rule tester for dry-run evaluation               |
| POST   | `/dashboard/test`   | Execute a dry-run test                           |

## Access control

The dashboard listens only on localhost by default. This provides basic security for development use, as only local processes can access it.

For production or shared environments, additional access control might be needed. Options include binding to a Unix socket only, requiring an API key via HTTP middleware, or integrating with an authentication proxy. These are future considerations.

## Responsive design

The dashboard is designed to be usable on smaller screens, though it's primarily a desktop developer tool. Tables scroll horizontally when needed. The event stream remains usable on mobile.

## Future enhancements

Possible future dashboard features include:

- Event search and filtering across historical events (requires event persistence)
- Rule editing with syntax highlighting and validation
- Configuration diffing between sources
- Performance profiling to identify slow rules
- Export/import of rules and state

These are not in the initial scope but could be added based on user feedback.

## See also

- [Architecture](../architecture.md) — high-level system architecture
- [Daemon](../daemon.md) — the long-running process that serves the dashboard
- [CLI dashboard command](../cli/dashboard.md) — starting the dashboard from the command line
