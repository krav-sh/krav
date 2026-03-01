# Real-time updates

The dashboard provides live updates through two mechanisms: WebSocket streaming for the event stream, and htmx polling for statistics and state views.

## WebSocket streaming

Live events use [`nhooyr.io/websocket`](https://pkg.go.dev/nhooyr.io/websocket) via the `/events` endpoint. The client subscribes and receives events as they occur. The server renders events as HTML fragments and pushes them to the client.

```go
import (
	"context"
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.CloseNow()

	ch := s.events.Subscribe()
	defer s.events.Unsubscribe(ch)

	ctx := conn.CloseRead(r.Context())

	// Render initial events as HTML
	initial := renderTemplate("event-list", &EventListData{Events: nil})
	if err := conn.Write(ctx, websocket.MessageText, initial); err != nil {
		return
	}

	// Stream new events as HTML fragments
	for {
		select {
		case event := <-ch:
			fragment := renderTemplate("event-item", &EventItemData{Event: event})
			if err := conn.Write(ctx, websocket.MessageText, fragment); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
```

The server uses a subscribe/unsubscribe pattern: each WebSocket connection registers a channel with the event bus, receives events as they're published, and unsubscribes on disconnect.

## htmx integration

The htmx WebSocket extension handles connecting to the WebSocket and swapping content into the DOM. New events are prepended to the event list using `hx-swap-oob="afterbegin"`:

```html
<div hx-ext="ws" ws-connect="/events">
    <div id="event-list" hx-swap-oob="afterbegin">
        <!-- Events are prepended here via WebSocket -->
    </div>
</div>
```

The server controls the rendering: it sends complete HTML fragments, and htmx inserts them. Client-side JavaScript does not need to parse data or build DOM elements.

## Polling

For statistics and state views, the dashboard uses htmx's polling feature rather than WebSocket. These views don't need instant updates; a 5-second refresh interval provides a good balance between freshness and server load:

```html
<div hx-get="/dashboard/stats" hx-trigger="every 5s" hx-swap="innerHTML">
    <!-- Stats content refreshed every 5 seconds -->
</div>
```

Polling is simpler than WebSocket for views that display aggregate data. The server renders the full view on each poll, so the client always gets a consistent snapshot.

## Configuration reload

Configuration status updates on file changes. When the daemon detects a configuration reload (via filesystem watcher), it publishes an event. The dashboard's config view either picks this up via the WebSocket connection or refreshes on its next poll cycle, ensuring users see the latest configuration state without manual refresh.
