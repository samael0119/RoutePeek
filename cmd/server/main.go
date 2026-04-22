package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/routepeek/internal/diagnostic"
	"github.com/routepeek/internal/discovery"
)

//go:embed static/*
var staticFiles embed.FS

var (
	port   = "8080"
	openBrowser = true
)

func main() {
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/snapshot", handleSnapshot).Methods("GET")
	r.HandleFunc("/api/diagnosis", handleDiagnosis).Methods("GET")
	r.HandleFunc("/api/interfaces", handleInterfaces).Methods("GET")
	r.HandleFunc("/api/routes", handleRoutes).Methods("GET")

	// Serve static files
	staticHandler := http.FileServer(http.FS(staticFiles))
	r.Handle("/", staticHandler)

	// Start server
	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("🌐 RoutePeek Web Server starting on http://localhost:%s\n", port)
	fmt.Printf("📊 Open http://localhost:%s in your browser\n", port)

	if openBrowser {
		go openBrowserFunc(fmt.Sprintf("http://localhost:%s", port))
	}

	srv := &http.Server{
		Handler: r,
		Addr:    addr,
	}

	log.Fatal(srv.ListenAndServe())
}

func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := discovery.GetNetworkSnapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func handleDiagnosis(w http.ResponseWriter, r *http.Request) {
	snapshot, err := discovery.GetNetworkSnapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	report := diagnostic.RunDiagnostics(snapshot)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func handleInterfaces(w http.ResponseWriter, r *http.Request) {
	interfaces, err := discovery.GetInterfaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(interfaces)
}

func handleRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := discovery.GetRoutes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

func openBrowserFunc(url string) {
	time.Sleep(500 * time.Millisecond)

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	}
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// StartWebServer starts the web server (called from CLI)
func StartWebServer(p string, open bool) {
	port = p
	openBrowser = open

	// Change to server directory context
	os.Chdir("/app")
	main()
}
