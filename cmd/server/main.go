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
	"github.com/routepeek/pkg/netinfo"
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
	r.HandleFunc("/api/trace", handleTrace).Methods("GET")
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFiles.ReadFile("static/favicon.png")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	})

	// Serve static files - strip "static" prefix from embedded FS
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles)))
	r.Handle("/static/", staticHandler)
	// Serve index.html at root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			data, err := staticFiles.ReadFile("static/index.html")
			if err != nil {
				http.Error(w, "index.html not found", 404)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
		} else {
			http.NotFound(w, r)
		}
	})

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

func handleTrace(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, `{"error":"target parameter is required"}`, http.StatusBadRequest)
		return
	}

	hops, err := netinfo.TraceRoute(target)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hops)
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
