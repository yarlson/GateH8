package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/yarlson/GateH8/config"
	"github.com/yarlson/GateH8/logger"
	"github.com/yarlson/GateH8/router"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

func main() {
	// Get the logger.
	log := logger.GetLogger()

	// Define the command-line argument for the server's address:port.
	var serverAddr string
	flag.StringVar(&serverAddr, "addr", ":1973", "Server address and port")
	flag.StringVar(&serverAddr, "a", ":1973", "Server address and port (shorthand)")

	// Customize the default flag.Usage function
	flag.Usage = Usage()

	flag.Parse()

	// Fetch the API Gateway's configuration using the utility function from the internal package.
	cfg, err := config.GetConfig()
	if err != nil {
		// Log and exit if there's an error loading the configuration.
		log.Fatal("Error loading configuration:", err)
	}

	// Initialize the router with the provided configuration. This router handles
	// requests based on the vhost, endpoint, and backend service configurations.
	r := router.NewRouter(cfg)

	// Create a new server and configure it.
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Use a channel to listen for interrupt signals to gracefully shutdown.
	done := make(chan struct{}, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	// This goroutine monitors the quit channel and gracefully shuts the server down.
	go func() {
		<-quit
		log.Info("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	// Log the start of the server and the port on which it is running.
	log.Infof("Server is ready to handle requests at %s", serverAddr)

	if cfg.UseTLS {
		srv.TLSConfig = &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				for vhostName, vhost := range cfg.Vhosts {
					if vhost.TLS != nil {
						// Using wildcard pattern matching to determine the appropriate certificate.
						match, err := filepath.Match(vhostName, info.ServerName)
						if err != nil {
							return nil, err
						}

						if match {
							cert, err := tls.LoadX509KeyPair(vhost.TLS.Cert, vhost.TLS.Key)
							if err != nil {
								return nil, err
							}
							return &cert, nil
						}
					}
				}
				return nil, fmt.Errorf("no certificate for given hostname: %s", info.ServerName)
			},
		}

		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting HTTPS server:", err)
		}
	} else {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting HTTP server:", err)
		}
	}

	<-done
	log.Info("Server stopped")
}

// Usage returns a function that prints the command-line usage message.
func Usage() func() {
	return func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Println("  -a, --addr string:   Server address and port (default \":1973\")")
		fmt.Println("  -h:                 Show this help message")
	}
}
