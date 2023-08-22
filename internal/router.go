package internal

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/hostrouter"
	"net/http"
)

// generateCORS creates a CORS middleware handler based on a given configuration.
// This is used to apply Cross-Origin Resource Sharing headers to responses, based on the provided configuration.
func generateCORS(c *CORSConfig) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   c.AllowedOrigins,
		AllowedMethods:   c.AllowedMethods,
		AllowedHeaders:   c.AllowedHeaders,
		ExposedHeaders:   c.ExposedHeaders,
		AllowCredentials: c.AllowCredentials,
		MaxAge:           c.MaxAge,
	})
}

// NewRouter constructs a new router based on a given configuration.
// The router manages incoming requests, directing them to the appropriate backend based on the requested host and path.
// Each virtual host (vhost) can have its own set of endpoints and CORS settings.
// Endpoints can additionally override the vhost's CORS settings if needed.
func NewRouter(config Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware layers to enrich request context and manage common API functionalities.
	r.Use(middleware.RequestID) // Assigns a unique ID to each request.
	r.Use(middleware.RealIP)    // Fetches the real IP from headers, useful if behind a proxy.
	r.Use(JsonLogger)           // A custom logger for logging request/response in JSON format.
	r.Use(middleware.Recoverer) // Recovers from panics and logs the stack trace.

	hr := hostrouter.New() // A router to manage routing based on request host (vhost).

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
			for _, method := range endpoint.Methods {
				endpointRouter.Method(method, endpoint.Path, CreateProxyHandler(endpoint.Backend))
			}
		}

		// Map the constructed vhost router to the corresponding host.
		hr.Map(vhost, router)
	}

	// Mount the host router to the main router.
	r.Mount("/", hr)
	return r
}
