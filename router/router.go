package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/yarlson/GateH8/config"
	"github.com/yarlson/GateH8/logger"
	"github.com/yarlson/GateH8/proxy"
	"net"
	"net/http"
	"path/filepath"
)

// generateCORS creates a CORS middleware handler based on a given configuration.
// This is used to apply Cross-Origin Resource Sharing headers to responses, based on the provided configuration.
func generateCORS(c *config.CORSConfig) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   c.AllowedOrigins,
		AllowedMethods:   c.AllowedMethods,
		AllowedHeaders:   c.AllowedHeaders,
		ExposedHeaders:   c.ExposedHeaders,
		AllowCredentials: c.AllowCredentials,
		MaxAge:           c.MaxAge,
	})
}

// WildcardHostRouter is a router that handles hostnames with wildcards and discards ports.
type WildcardHostRouter struct {
	routes map[string]*chi.Mux
}

// NewWildcardHostRouter initializes a new WildcardHostRouter.
func NewWildcardHostRouter() *WildcardHostRouter {
	return &WildcardHostRouter{
		routes: make(map[string]*chi.Mux),
	}
}

// Map maps a host pattern to a router.
func (whr *WildcardHostRouter) Map(pattern string, router *chi.Mux) {
	whr.routes[pattern] = router
}

// Route routes based on host patterns.
func (whr *WildcardHostRouter) Route(w http.ResponseWriter, r *http.Request) {
	host, _, _ := net.SplitHostPort(r.Host)
	if host == "" {
		host = r.Host // in case SplitHostPort failed, which means there was no port
	}
	for pattern, router := range whr.routes {
		if matched, _ := filepath.Match(pattern, host); matched {
			router.ServeHTTP(w, r)
			return
		}
	}
	http.Error(w, "Host not found", http.StatusNotFound)
}

// Handler is a http.Handler implementation of Route.
func (whr *WildcardHostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	whr.Route(w, r)
}

// NewRouter constructs a new router based on a given configuration.
// The router manages incoming requests, directing them to the appropriate backend based on the requested host and path.
// Each virtual host (vhost) can have its own set of endpoints and CORS settings.
// Endpoints can additionally override the vhost's CORS settings if needed.
func NewRouter(config *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware layers to enrich request context and manage common API functionalities.
	r.Use(middleware.RequestID) // Assigns a unique ID to each request.
	r.Use(middleware.RealIP)    // Fetches the real IP from headers, useful if behind a proxy.
	r.Use(logger.JsonLogger)    // A custom logger for logging request/response in JSON format.
	r.Use(middleware.Recoverer) // Recovers from panics and logs the stack trace.

	hr := NewWildcardHostRouter() // A router to manage routing based on request host (vhost).

	// Iterate over each virtual host in the configuration.
	for vhost, vhostConfig := range config.Vhosts {
		router := chi.NewRouter()

		// Apply vhost level CORS if specified.
		if vhostConfig.CORS != nil {
			router.Use(generateCORS(vhostConfig.CORS))
		}

		// Set up each endpoint for the virtual host.
		for _, endpoint := range vhostConfig.Endpoints {
			endpointRouter := router

			// If an endpoint has specific CORS settings, we override the vhost CORS.
			if endpoint.CORS != nil {
				endpointRouter = chi.NewRouter()
				if vhostConfig.CORS != nil {
					endpointRouter.Use(generateCORS(vhostConfig.CORS)) // Reapply vhost CORS before endpoint CORS.
				}
				endpointRouter.Use(generateCORS(endpoint.CORS)) // Apply specific endpoint CORS.
				router.Mount(endpoint.Path, endpointRouter)
			}

			// Bind all the allowed methods for the endpoint to the respective handler.
			if endpoint.WebSocket != nil {
				endpointRouter.HandleFunc(endpoint.Path, proxy.CreateWebSocketProxyHandler(endpoint))
			} else {
				for _, method := range endpoint.Methods {
					endpointRouter.Method(method, endpoint.Path, proxy.CreateHttpProxyHandler(endpoint.Backend, &http.Client{}))
				}
			}
		}

		// Map the constructed vhost router to the corresponding host.
		hr.Map(vhost, router)
	}

	// Mount the host router to the main router.
	r.Mount("/", hr)
	return r
}
