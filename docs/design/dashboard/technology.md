# Dashboard technology

The dashboard uses a server-side rendering approach with minimal client-side JavaScript. This keeps the frontend simple, avoids JavaScript build complexity, and provides a responsive experience without a full SPA system.

## Server-side rendering

The dashboard uses Go's `html/template` with [Sprig](https://masterminds.github.io/sprig/) functions for server-side rendering. Handlers pass template data as typed structs, and `html/template` automatically escapes output for XSS protection (unlike `text/template`, which does no escaping).

```go
func (s *Server) eventsPage(w http.ResponseWriter, r *http.Request) {
    filters := parseEventFilters(r)
    events := s.events.Recent(100)
    data := EventsPageData{
        Events:  events,
        Filters: filters,
    }
    s.templates.ExecuteTemplate(w, "events.html", data)
}
```

## Template organization

Templates live in a `templates/` directory, and the server parses them at startup. The directory follows a standard layout:

- **Base layout**: shared HTML structure (head, nav, footer)
- **Page templates**: one per dashboard view (events, stats, config, state, test)
- **Partials**: reusable fragments (event item, rule row, config source)

Templates compose through Go's `{{ template }}` and `{{ block }}` actions, with the base layout defining blocks that page templates override.

## htmx

[htmx](https://htmx.org/) provides client-side interactivity. Instead of writing JavaScript, htmx attributes on HTML elements trigger HTTP requests and swap content. The result is an interactive experience with minimal client-side code.

Key patterns used in the dashboard:

- `hx-get`: fetch and swap page fragments without full reloads
- `hx-swap`: control where fetched content goes (innerHTML, outerHTML, afterbegin)
- `hx-trigger`: trigger requests on events or intervals (such as `every 5s` for polling)
- `hx-ext="ws"`: WebSocket extension for live event streaming

The team chose htmx over a SPA system because the dashboard's interactivity needs are modest: swapping content fragments, polling for updates, and streaming events. A full SPA would add build complexity and JavaScript dependencies without proportional benefit.

## Alpine.js

[Alpine.js](https://alpinejs.dev/) handles small bits of client-side state where htmx alone isn't sufficient: toggle switches, filter dropdowns, expand/collapse controls. It complements htmx by managing UI state that doesn't need a server round-trip.

## CSS library

Styling uses [Pico CSS](https://picocss.com/) (or similar minimal library) for clean defaults without heavy dependencies. Pico provides:

- Classless styling: semantic HTML elements receive styles automatically
- Responsive defaults: layouts adapt to screen size without custom breakpoints
- Minimal footprint: no build step, no utility classes to learn

This keeps the dashboard visually clean while avoiding CSS library lock-in.
