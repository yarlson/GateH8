package main

import (
	"github.com/yarlson/GateH8/internal"
	"net/http"
)

func main() {
	// Fetch the API Gateway's configuration using the utility function from the internal package.
	err, config := internal.GetConfig()
	if err != nil {
		// Log and exit if there's an error loading the configuration.
		internal.L.Fatal("Error loading configuration:", err)
	}

	// Initialize the router with the provided configuration. This router handles
	// requests based on the vhost, endpoint, and backend service configurations.
	r := internal.NewRouter(config)

	// Log the start of the server and the port on which it is running.
	internal.L.Info("Starting server on port :1973")

	// Start the HTTP server on port 1973 and bind it to the configured router.
	err = http.ListenAndServe(":1973", r)
	if err != nil {
		// Log and exit if there's an error starting the HTTP server.
		internal.L.Fatal("Error starting server:", err)
	}
}
