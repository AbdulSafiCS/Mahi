package main

import (
	"fmt"
	"net/http"
	"os"

	"mahi/server/internal/config"
	httpserver "mahi/server/internal/http"
)

//
func main() {
	//read env variables and return struct
	cfg := config.Load()
	//builds the router with all http routes
	r := httpserver.NewRouter(cfg)
	//starts the http server
	addr := fmt.Sprintf(":%s", cfg.Port)
	fmt.Printf("API listening on %s\n", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}
