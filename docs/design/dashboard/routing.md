# Dashboard routing

The dashboard organizes routes as a chi subrouter that nests under `/dashboard`. Each page has a corresponding handler that fetches data from the daemon's shared state and renders a Go template.

## Router setup

The dashboard registers its routes as a group using [chi](https://github.com/go-chi/chi):

```go
import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func dashboardRoutes(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.handleDashboardIndex)
	r.Get("/events", s.handleEventsPage)
	r.Get("/stats", s.handleStatsPage)
	r.Get("/config", s.handleConfigPage)
	r.Get("/state", s.handleStatePage)
	r.Get("/test", s.handleTestPage)
	r.Post("/test", s.handleRunTest)
	return r
}

func (s *Server) handleDashboardIndex(w http.ResponseWriter, r *http.Request) {
	recentEvents := s.events.Recent(50)
	renderPage(w, "dashboard.html", &DashboardData{Events: recentEvents})
}
```

The daemon's main router mounts the returned `http.Handler` at `/dashboard`, keeping dashboard routes cleanly separated from the API.

## Handler pattern

Each handler follows a consistent pattern:

1. **Parse request**: extract query parameters, filters, or form data from the request
2. **Fetch data**: read from the daemon's cached state (events, rules, config, state store)
3. **Build struct**: assemble a typed data struct for the template
4. **Render template**: execute the named template with the data struct

This pattern keeps handlers simple and testable, as each step is a straightforward function call.

## Data source

The dashboard reads from the daemon's cached state rather than hitting the filesystem or SQLite directly. The dashboard shows the same view of configuration and state that evaluation uses. The daemon updates its cache on configuration reloads and state changes, so the dashboard always reflects current state without its own cache invalidation logic.

## Adding a new page

To add a new dashboard page:

1. **Create the template**: add a new `.html` file in `templates/` extending the base layout
2. **Define the data struct**: create a typed struct for the template's data
3. **Write the handler**: follow the parse, fetch, struct, render pattern
4. **Register the route**: add a `r.Get(...)` or `r.Post(...)` line in `dashboardRoutes`
5. **Add navigation**: include a link in the base layout's nav partial
6. **Update the route table**: add the route to the [index](index.md#routes)
